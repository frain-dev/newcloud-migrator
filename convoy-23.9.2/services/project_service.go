package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/auth"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/config"
	"github.com/oklog/ulid/v2"

	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/api/models"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/cache"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/datastore"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/limiter"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/pkg/log"
	"github.com/frain-dev/newcloud-migrator/convoy-23.9.2/util"
)

type ProjectService struct {
	apiKeyRepo        datastore.APIKeyRepository
	projectRepo       datastore.ProjectRepository
	eventRepo         datastore.EventRepository
	eventDeliveryRepo datastore.EventDeliveryRepository
	limiter           limiter.RateLimiter
	cache             cache.Cache
}

func NewProjectService(apiKeyRepo datastore.APIKeyRepository, projectRepo datastore.ProjectRepository, eventRepo datastore.EventRepository, eventDeliveryRepo datastore.EventDeliveryRepository, cache cache.Cache) (*ProjectService, error) {
	cfg, err := config.Get()
	if err != nil {
		return nil, err
	}

	rlimiter, err := limiter.NewLimiter(cfg.Redis)
	if err != nil {
		return nil, err
	}

	return &ProjectService{
		apiKeyRepo:        apiKeyRepo,
		projectRepo:       projectRepo,
		eventRepo:         eventRepo,
		eventDeliveryRepo: eventDeliveryRepo,
		limiter:           rlimiter,
		cache:             cache,
	}, nil
}

func (ps *ProjectService) CreateProject(ctx context.Context, newProject *models.CreateProject, org *datastore.Organisation, member *datastore.OrganisationMember) (*datastore.Project, *models.APIKeyResponse, error) {
	projectName := newProject.Name

	projectConfig := newProject.Config.Transform()
	if projectConfig == nil {
		projectConfig = &datastore.DefaultProjectConfig
	} else {
		if projectConfig.Signature != nil {
			checkSignatureVersions(projectConfig.Signature.Versions)
		} else {
			projectConfig.Signature = datastore.DefaultProjectConfig.Signature
		}

		if projectConfig.RateLimit == nil {
			projectConfig.RateLimit = datastore.DefaultProjectConfig.RateLimit
		}

		if projectConfig.Strategy == nil {
			projectConfig.Strategy = datastore.DefaultProjectConfig.Strategy
		}

		err := validateMetaEvent(projectConfig.MetaEvent)
		if err != nil {
			return nil, nil, util.NewServiceError(http.StatusBadRequest, err)
		}

		if projectConfig.RetentionPolicy != nil {
			if !util.IsStringEmpty(projectConfig.RetentionPolicy.SearchPolicy) {
				_, err = time.ParseDuration(projectConfig.RetentionPolicy.SearchPolicy)
				if err != nil {
					return nil, nil, util.NewServiceError(http.StatusBadRequest, err)
				}
			}

			if !util.IsStringEmpty(projectConfig.RetentionPolicy.Policy) {
				_, err = time.ParseDuration(projectConfig.RetentionPolicy.Policy)
				if err != nil {
					return nil, nil, util.NewServiceError(http.StatusBadRequest, err)
				}
			}
		}
	}

	project := &datastore.Project{
		UID:            ulid.Make().String(),
		Name:           projectName,
		Type:           datastore.ProjectType(newProject.Type),
		OrganisationID: org.UID,
		Config:         projectConfig,
		LogoURL:        newProject.LogoURL,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := ps.projectRepo.CreateProject(ctx, project)
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("failed to create project")
		if errors.Is(err, datastore.ErrDuplicateProjectName) {
			return nil, nil, util.NewServiceError(http.StatusBadRequest, err)
		}

		return nil, nil, util.NewServiceError(http.StatusBadRequest, errors.New("failed to create project"))
	}

	newAPIKey := &models.APIKey{
		Name: fmt.Sprintf("%s's default key", project.Name),
		Role: models.Role{
			Type:    auth.RoleAdmin,
			Project: project.UID,
		},
	}

	cak := CreateAPIKeyService{
		ProjectRepo: ps.projectRepo,
		APIKeyRepo:  ps.apiKeyRepo,
		Member:      member,
		NewApiKey:   newAPIKey,
	}

	apiKey, keyString, err := cak.Run(ctx)
	if err != nil {
		return nil, nil, err
	}

	resp := &models.APIKeyResponse{
		APIKey: models.APIKey{
			Name: apiKey.Name,
			Role: models.Role{
				Type:    apiKey.Role.Type,
				Project: apiKey.Role.Project,
			},
			Type:      apiKey.Type,
			ExpiresAt: apiKey.ExpiresAt,
		},
		UID:       apiKey.UID,
		CreatedAt: apiKey.CreatedAt,
		Key:       keyString,
	}

	return project, resp, nil
}

func (ps *ProjectService) UpdateProject(ctx context.Context, project *datastore.Project, update *models.UpdateProject) (*datastore.Project, error) {
	if !util.IsStringEmpty(update.Name) {
		project.Name = update.Name
	}

	if update.Config != nil {
		if update.Config.RetentionPolicy != nil {
			if !util.IsStringEmpty(update.Config.RetentionPolicy.SearchPolicy) {
				_, err := time.ParseDuration(update.Config.RetentionPolicy.SearchPolicy)
				if err != nil {
					return nil, util.NewServiceError(http.StatusBadRequest, err)
				}
			}

			if !util.IsStringEmpty(update.Config.RetentionPolicy.Policy) {
				_, err := time.ParseDuration(update.Config.RetentionPolicy.Policy)
				if err != nil {
					return nil, util.NewServiceError(http.StatusBadRequest, err)
				}
			}
		}

		project.Config = update.Config.Transform()
		checkSignatureVersions(project.Config.Signature.Versions)
		err := validateMetaEvent(project.Config.MetaEvent)
		if err != nil {
			return nil, util.NewServiceError(http.StatusBadRequest, err)
		}
	}

	if !util.IsStringEmpty(update.LogoURL) {
		project.LogoURL = update.LogoURL
	}

	err := ps.projectRepo.UpdateProject(ctx, project)
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("failed to to update project")
		return nil, util.NewServiceError(http.StatusBadRequest, err)
	}

	projectCacheKey := convoy.ProjectsCacheKey.Get(project.UID).String()
	err = ps.cache.Set(ctx, projectCacheKey, &project, time.Minute*5)
	if err != nil {
		return nil, util.NewServiceError(http.StatusBadRequest, err)
	}

	return project, nil
}

func checkSignatureVersions(versions []datastore.SignatureVersion) {
	for i := range versions {
		v := &versions[i]
		if v.UID == "" {
			v.UID = ulid.Make().String()
		}

		if v.CreatedAt.Unix() == 0 {
			v.CreatedAt = time.Now()
		}
	}
}

func validateMetaEvent(metaEvent *datastore.MetaEventConfiguration) error {
	if metaEvent == nil {
		return nil
	}

	if !metaEvent.IsEnabled {
		return nil
	}

	if metaEvent.Type == datastore.HTTPMetaEvent {
		url, err := util.CleanEndpoint(metaEvent.URL)
		if err != nil {
			return err
		}
		metaEvent.URL = url
	}

	if util.IsStringEmpty(metaEvent.Secret) {
		sc, err := util.GenerateSecret()
		if err != nil {
			return err
		}

		metaEvent.Secret = sc
	}

	return nil
}
