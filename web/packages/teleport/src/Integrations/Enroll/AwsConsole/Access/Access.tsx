/**
 * Teleport
 * Copyright (C) 2025 Gravitational, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { useCallback, useEffect, useState } from 'react';
import { useHistory, useLocation, useParams } from 'react-router';

import * as Alerts from 'design/Alert';
import { Alert } from 'design/Alert';
import Box from 'design/Box';
import { ButtonPrimary, ButtonSecondary } from 'design/Button';
import Flex from 'design/Flex';
import { Indicator } from 'design/Indicator';
import { H2, P2 } from 'design/Text';
import { InfoGuideButton } from 'shared/components/SlidingSidePanel/InfoGuide';

import { EmptyState } from 'teleport/Bots/List/EmptyState/EmptyState';
import { FeatureBox } from 'teleport/components/Layout';
import cfg from 'teleport/config';
import { Profiles } from 'teleport/Integrations/Enroll/AwsConsole/Access/Profiles';
import { ProfilesEmptyState } from 'teleport/Integrations/Enroll/AwsConsole/Access/ProfilesEmptyState';
import { ProfilesFilterOption } from 'teleport/Integrations/Enroll/AwsConsole/Access/ProfilesFilter';
import { Guide } from 'teleport/Integrations/Enroll/AwsConsole/Guide';
import {
  IntegrationKind,
  integrationService,
} from 'teleport/services/integrations';
import useTeleport from 'teleport/useTeleport';

export function Access() {
  const ctx = useTeleport();
  const flags = ctx.getFeatureFlags();
  // todo mberg all layers
  // todo mberg is correct?
  const canEnroll = flags.enrollIntegrations;
  const clusterId = ctx.storeUser.getClusterId();
  const resourceRoute = cfg.getUnifiedResourcesRoute(clusterId);

  const history = useHistory();
  const location = useLocation<{ prevPageTokens?: readonly string[] }>();
  const queryParams = new URLSearchParams(location.search);
  const pageToken = queryParams.get('page') ?? '';
  const sort = queryParams.get('sort') || 'active_at_latest:desc';

  const { name } = useParams<{ name: string }>();
  const [syncAll, setSyncAll] = useState(true);
  const [filters, setFilters] = useState<ProfilesFilterOption[]>([]);

  const { status, error, data, isFetching, refetch } = useQuery({
    enabled: canEnroll,
    queryKey: ['profiles', filters, pageToken, sort],
    queryFn: () =>
      integrationService.awsRolesAnywhereProfiles({
        integrationName: name,
        pageSize: 20,
        pageToken,
        filters,
      }),
    placeholderData: keepPreviousData,
    staleTime: 30_000, // Cached pages are valid for 30 seconds
  });

  const { prevPageTokens = [] } = location.state ?? {};
  const hasNextPage = !!data?.nextPageToken;
  const hasPrevPage = !!pageToken;

  const handleFetchNext = useCallback(() => {
    const search = new URLSearchParams(location.search);
    search.set('page', data?.nextPageToken ?? '');

    history.replace(
      {
        pathname: location.pathname,
        search: search.toString(),
      },
      {
        prevPageTokens: [...prevPageTokens, pageToken],
      }
    );
  }, [
    data?.nextPageToken,
    history,
    location.pathname,
    location.search,
    pageToken,
    prevPageTokens,
  ]);

  const handleFetchPrev = useCallback(() => {
    const prevTokens = [...prevPageTokens];
    const nextToken = prevTokens.pop();

    const search = new URLSearchParams(location.search);
    search.set('page', nextToken ?? '');

    history.replace(
      {
        pathname: location.pathname,
        search: search.toString(),
      },
      {
        prevPageTokens: prevTokens,
      }
    );
  }, [history, location.pathname, location.search, prevPageTokens]);

  useEffect(() => {
    // clear filters on syncAll
    if (syncAll) {
      setFilters([]);
    }
  }, [syncAll]);

  // todo mberg get role
  if (!canEnroll) {
    return (
      <FeatureBox>
        <Alert kind="info" mt={4}>
          You do not have permission to enroll integrations. Missing role
          permissions: <code>todo</code>
        </Alert>
        <EmptyState />
      </FeatureBox>
    );
  }

  return (
    <Box pt={3}>
      <Flex mt={3} justifyContent="space-between" alignItems="center">
        <H2>Configure Access</H2>
        <InfoGuideButton
          config={{
            guide: <Guide resourcesRoute={resourceRoute} />,
          }}
        />
      </Flex>
      <P2 mb={3}>
        Import and synchronize AWS IAM Roles Anywhere Profiles into Teleport.
        Imported Profiles will be available as Resources with each Role
        available as an account.
      </P2>
      {status === 'error' && (
        <Alerts.Danger details={error.message}>
          Error: {error.name}
        </Alerts.Danger>
      )}
      {status === 'pending' && (
        <Box data-testid="loading" textAlign="center" m={10}>
          <Indicator />
        </Box>
      )}
      {status === 'success' &&
        (!data.profiles || data.profiles.length === 0) && (
          <ProfilesEmptyState />
        )}
      {status === 'success' && data.profiles.length > 0 && (
        <Profiles
          data={data.profiles}
          fetchStatus={isFetching ? 'loading' : ''}
          filters={filters}
          setFilters={setFilters}
          onFetchNext={hasNextPage ? handleFetchNext : undefined}
          onFetchPrev={hasPrevPage ? handleFetchPrev : undefined}
          refetch={refetch}
          syncAll={syncAll}
          setSyncAll={setSyncAll}
        />
      )}
      <Flex gap={3} mt={3}>
        <ButtonPrimary
          width="200px"
          disabled={!syncAll || filters.length !== 0}
        >
          Enable Sync
        </ButtonPrimary>
        <ButtonSecondary
          onClick={() =>
            history.push(
              cfg.getIntegrationEnrollChildRoute(
                IntegrationKind.AwsOidc,
                name,
                IntegrationKind.AwsConsole,
                'integration'
              )
            )
          }
          width="100px"
        >
          Back
        </ButtonSecondary>
      </Flex>
    </Box>
  );
}
