package main

import (
	"context"
	"encoding/json"
	"fmt"
	ncache "github.com/frain-dev/convoy/cache/noop"
	"github.com/frain-dev/convoy/database/postgres"
	"github.com/frain-dev/convoy/datastore"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

type pagedResponse struct {
	Data data `json:"data"`
}

type data struct {
	Content    interface{}              `json:"content,omitempty"`
	Pagination datastore.PaginationData `json:"pagination,omitempty"`
}

func (m *Migrator) loadUser() (*datastore.User, error) {
	url := fmt.Sprintf("%s/ui/users/random/profile", m.OldBaseURL)
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	m.addHeader(r)

	resp, err := client.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	type data struct {
		Data interface{} `json:"data,omitempty"`
	}

	user := &datastore.User{}

	err = readBody(resp.Body, &data{user})
	if err != nil {
		return nil, fmt.Errorf("failed to read user body: %v", err)
	}

	return user, nil
}

func (m *Migrator) loadOrganisations(pageable pagedResponse) ([]datastore.Organisation, error) {
	url := fmt.Sprintf("%s/ui/organisations?perPage=%d&direction=next&next_page_cursor=%s", m.OldBaseURL, pageable.Data.Pagination.PerPage, pageable.Data.Pagination.NextPageCursor)
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	m.addHeader(r)

	resp, err := client.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to get orgs: %v", err)
	}

	orgs := []datastore.Organisation{}
	pg := pagedResponse{
		Data: data{
			Content: &orgs,
		},
	}

	err = readBody(resp.Body, &pg)
	if err != nil {
		return nil, fmt.Errorf("failed to read orgs body: %v", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("request failed: %d", resp.StatusCode)
	}

	if pg.Data.Pagination.HasNextPage {
		moreOrgs, err := m.loadOrganisations(pg)
		if err != nil {
			log.WithError(err).Errorf("failed to load next org page, next cursor is %s", pg.Data.Pagination.NextPageCursor)
		}

		orgs = append(orgs, moreOrgs...)
	}

	return orgs, nil
}

func readBody(r io.ReadCloser, i interface{}) error {
	defer r.Close()

	b, err := io.ReadAll(r)
	r.Close()
	if err != nil {
		return err
	}

	fmt.Println("body", string(b))
	return json.Unmarshal(b, i)
}

func (m *Migrator) loadOrgProjects(projectRepo datastore.ProjectRepository, orgID string) ([]*datastore.Project, error) {
	return projectRepo.LoadProjects(context.Background(), &datastore.ProjectFilter{OrgID: orgID})
}

func (m *Migrator) loadProjectEndpoints(endpointRepo datastore.EndpointRepository, projectID string, pageable datastore.Pageable) ([]datastore.Endpoint, error) {
	endpoints, paginationData, err := endpointRepo.LoadEndpointsPaged(context.Background(), projectID, &datastore.Filter{}, pageable)
	if err != nil {
		return nil, err
	}

	if paginationData.HasNextPage {
		pageable.NextCursor = endpoints[len(endpoints)-1].UID
		moreEndpoints, err := m.loadProjectEndpoints(endpointRepo, projectID, pageable)
		if err != nil {
			log.WithError(err).Errorf("failed to load next members page, next cursor is %s", paginationData.NextPageCursor)
		}

		endpoints = append(endpoints, moreEndpoints...)
	}

	return endpoints, nil
}

func (m *Migrator) addHeader(r *http.Request) {
	r.Header.Add("Authorization", "Bearer "+m.PAT)
}

func (m *Migrator) loadProjectSources(sourceRepo datastore.SourceRepository, projectID string, pageable datastore.Pageable) error {
	sources, paginationData, err := sourceRepo.LoadSourcesPaged(context.Background(), projectID, &datastore.SourceFilter{}, pageable)
	if err != nil {
		return err
	}

	if len(sources) > 0 {
		err = m.SaveSources(context.Background(), sources)
		if err != nil {
			return fmt.Errorf("failed to save sources: %v", err)
		}
	}

	for _, source := range sources {
		m.sourceIDs[source.UID] = struct{}{}
	}

	if paginationData.HasNextPage {
		pageable.NextCursor = sources[len(sources)-1].UID
		err := m.loadProjectSources(sourceRepo, projectID, pageable)
		if err != nil {
			log.WithError(err).Errorf("failed to load next members page, next cursor is %s", paginationData.NextPageCursor)
		}
	}

	return nil
}

func (m *Migrator) loadProjectSubscriptions(subRepo datastore.SubscriptionRepository, projectID string, pageable datastore.Pageable) error {
	subscriptions, paginationData, err := subRepo.LoadSubscriptionsPaged(context.Background(), projectID, &datastore.FilterBy{}, pageable)
	if err != nil {
		return err
	}

	err = m.SaveSubscriptions(context.Background(), subscriptions)
	if err != nil {
		return fmt.Errorf("failed to save subscriptions: %v", err)
	}

	for _, subscription := range subscriptions {
		m.subIDs[subscription.UID] = struct{}{}
	}

	if paginationData.HasNextPage {
		pageable.NextCursor = subscriptions[len(subscriptions)-1].UID
		err := m.loadProjectSubscriptions(subRepo, projectID, pageable)
		if err != nil {
			log.WithError(err).Errorf("failed to load next members page, next cursor is %s", paginationData.NextPageCursor)
		}

	}

	return nil
}

func (m *Migrator) loadAPIKeys(apiKeyRepo datastore.APIKeyRepository, projectID, userID string, pageable datastore.Pageable) error {
	f := &datastore.ApiKeyFilter{
		ProjectID: projectID,
	}

	if userID != "" {
		f = &datastore.ApiKeyFilter{
			UserID: userID,
		}
	}

	keys, paginationData, err := apiKeyRepo.LoadAPIKeysPaged(context.Background(), f, &pageable)
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		err = m.SaveAPIKeys(context.Background(), keys)
		if err != nil {
			return fmt.Errorf("failed to save project keys: %v", err)
		}
	}

	if paginationData.HasNextPage {
		pageable.NextCursor = keys[len(keys)-1].UID
		err := m.loadAPIKeys(apiKeyRepo, projectID, userID, pageable)
		if err != nil {
			log.WithError(err).Errorf("failed to load next api keys page, next cursor is %s", paginationData.NextPageCursor)
		}
	}

	return nil
}

func (m *Migrator) loadOrgMembers(orgMemberRepo datastore.OrganisationMemberRepository, orgID string, pageable datastore.Pageable) ([]*datastore.OrganisationMember, error) {
	members, paginationData, err := orgMemberRepo.LoadOrganisationMembersPaged(context.Background(), orgID, "", pageable)
	if err != nil {
		return nil, err
	}

	if paginationData.HasNextPage {
		pageable.NextCursor = members[len(members)-1].UID
		moreMembers, err := m.loadOrgMembers(orgMemberRepo, orgID, pageable)
		if err != nil {
			log.WithError(err).Errorf("failed to load next members page, next cursor is %s", paginationData.NextPageCursor)
		}

		members = append(members, moreMembers...)
	}

	return members, nil
}

func (m *Migrator) loadUsers(userIDs map[string]struct{}) ([]*datastore.User, error) {
	var users []*datastore.User
	userRepo := postgres.NewUserRepo(m, ncache.NewNoopCache())

	for userID := range userIDs {
		user, err := userRepo.FindUserByID(context.Background(), userID)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (m *Migrator) loadEvents(eventRepo datastore.EventRepository, project *datastore.Project, pageable datastore.Pageable) error {
	f := &datastore.Filter{
		Project:  project,
		Pageable: pageable,
		SearchParams: datastore.SearchParams{
			CreatedAtStart: 0,
			CreatedAtEnd:   time.Now().Unix(),
		},
	}

	events, paginationData, err := eventRepo.LoadEventsPaged(context.Background(), project.UID, f)
	if err != nil {
		return err
	}

	if len(events) > 0 {
		err = m.SaveEvents(context.Background(), events)
		if err != nil {
			return fmt.Errorf("failed to save events: %v", err)
		}

		for _, event := range events {
			m.eventIDs[event.UID] = struct{}{}
		}
	}

	if paginationData.HasNextPage {
		f.Pageable.NextCursor = events[len(events)-1].UID
		err := m.loadEvents(eventRepo, project, f.Pageable)
		if err != nil {
			log.WithError(err).Errorf("failed to load next event page, next cursor is %s", paginationData.NextPageCursor)
		}
	}

	return nil
}

func (m *Migrator) loadEventDeliveries(eventDeliveryRepository datastore.EventDeliveryRepository, project *datastore.Project, pageable datastore.Pageable) error {
	eventDeliveries, paginationData, err := eventDeliveryRepository.LoadEventDeliveriesPaged(
		context.Background(),
		project.UID, nil, "", "", nil, datastore.SearchParams{
			CreatedAtStart: 0,
			CreatedAtEnd:   time.Now().Unix(),
		}, pageable, "",
	)
	if err != nil {
		return err
	}

	if len(eventDeliveries) > 0 {
		err = m.SaveEventDeliveries(context.Background(), eventDeliveries)
		if err != nil {
			return fmt.Errorf("failed to save deliveries: %v", err)
		}
	}

	if paginationData.HasNextPage {
		pageable.NextCursor = eventDeliveries[len(eventDeliveries)-1].UID
		err := m.loadEventDeliveries(eventDeliveryRepository, project, pageable)
		if err != nil {
			log.WithError(err).Errorf("failed to load next event page, next cursor is %s", paginationData.NextPageCursor)
		}

	}

	return nil
}
