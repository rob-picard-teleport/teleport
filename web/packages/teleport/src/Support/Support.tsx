/**
 * Teleport
 * Copyright (C) 2023  Gravitational, Inc.
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

import React from 'react';
import { Link } from 'react-router-dom';
import styled from 'styled-components';

import { Box, Flex, H2, H3, Text } from 'design';
import * as Icons from 'design/Icon';
import { MultiRowBox, Row } from 'design/MultiRowBox';

import { ButtonLockedFeature } from 'teleport/components/ButtonLockedFeature';
import {
  FeatureBox,
  FeatureHeader,
  FeatureHeaderTitle,
} from 'teleport/components/Layout';
import cfg from 'teleport/config';
import { useNoMinWidth } from 'teleport/Main';
import { CtaEvent } from 'teleport/services/userEvent';
import useTeleport from 'teleport/useTeleport';

import api from 'teleport/services/api/api';
import Table, { Cell } from 'design/DataTable';



export function SupportContainer({ children }: { children?: React.ReactNode }) {
  const ctx = useTeleport();
  const cluster = ctx.storeUser.state.cluster;

  // showCTA returns the premium support value for enterprise customers and true for OSS users
  const showCTA = cfg.isEnterprise ? !cfg.premiumSupport : true;

  return (
    <Support
      {...cluster}
      isEnterprise={cfg.isEnterprise}
      tunnelPublicAddress={cfg.tunnelPublicAddress}
      isCloud={cfg.isCloud}
      showPremiumSupportCTA={showCTA}
      children={children}
    />
  );
}

export const Support = ({
  clusterId,
  authVersion,
  publicURL,
  isEnterprise,
  licenseExpiryDateText,
  tunnelPublicAddress,
  isCloud,
  children,
  showPremiumSupportCTA,
  serviceItems = [   

     {
        "Name": "ssh.node",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "proxy",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "proxy.web",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "fileuploadcompleter.service",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "common.upload.init",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "fileuploader.shutdown",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "auth.tls",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "proxy",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "readyz.monitor",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "fileuploadcompleter.shutdown",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "proxy.db.postgres",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "proxy.reversetunnel.watcher",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "auth.shutdown",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "proxy.shutdown",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "proxy.reversetunnel.web",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "register.proxy",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "auth.expiry",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "proxy.reversetunnel.tls",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "register.node",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "auth.broadcast",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "register.instance",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "proxy.reversetunnel.server",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "auth.server.system-clock-monitor",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "proxy.tls.alpn.sni.proxy.reverseTunnel",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "fileuploader.service",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "debug.shutdown",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "proxy.tls.alpn.sni.proxy",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "closer",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "auth.heartbeat",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "auth.autoupdate_agent_rollout_controller",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "debug.service",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "ssh.node",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "common.rotate",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "auth.heartbeat.broadcast",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "auth",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-zxc",
        "Hostname": "im-marcos-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "proxy.ssh",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "enterprise.auth.services.stop",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "instance.init",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "proxy.init",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "idp.saml",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "auth.server_info",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "proxy.db.tls",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "auth.spiffe_federation_syncer",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "node",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "update.aws-oidc.deploy.service",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    },
    {
        "Name": "proxy.grpc.public",
        "Status": "Running",
        "HostID": "269d1652-8fdd-47ff-bf6d-09a3e1f51ba6",
        "Hostname": "im-a-nodename",
        "ProcessUptime": "1 hour ago",
        "Version": "18.0.0-dev"
    }
  ],
}: Props) => {
  useNoMinWidth();
  const docs = getDocUrls(authVersion, isEnterprise);


// async function fetchProcessHealth() {
//   const url = cfg.getProcessHealthUrl();
//   try {
//     const response = await fetch(url, {
//       credentials: 'include', // include cookies if needed for auth
//     });
//     if (!response.ok) {
//       throw new Error(`HTTP error! status: ${response.status}`);
//     }
//     const data = await response.json();
//     console.log('Process Health:', data);
//     return data;
//   } catch (err) {
//     console.error('Failed to fetch process health:', err);
//     return null;
//   }
// }

//  fetchProcessHealth();


// ...existing code...

// ** TODO (MP/mayurah) ** : Refactor this to use the api.get method for consistent headers and error handling

async function fetchProcessHealth() {
  const url = cfg.getProcessHealthUrl();
  try {
    // Use the api.get method for consistent headers and error handling
    const data = await api.get(url);
    // console.log('Process Health:', data);

    // const serviceItems = [];

    data.Items.forEach(item => {
      const { HostID, Hostname, ProcessUptime, Version, Units } = item;
      Units.forEach(unit => {
        serviceItems.push({
          ...unit,
          HostID,
          Hostname,
          ProcessUptime,
          Version,
        });
      });
    });

    console.log('Service Items:', serviceItems);

    /**
     * To trigger a re-render after updating serviceItems, you need to manage serviceItems as a state variable.
     * Move serviceItems into a useState, and update it here.
     * Example:
     */

    return data;
  } catch (err) {
    console.error('Failed to fetch process health:', err);
    return null;
  }
}


const processHealth = fetchProcessHealth().then(data => {
  if (data) {
    console.log('Process Health:', data);
  }
});


  return (
    <FeatureBox maxWidth="2000px">
      <FeatureHeader>
        <FeatureHeaderTitle>Help & Support</FeatureHeaderTitle>
      </FeatureHeader>
      <StyledMultiRowBox mb={3}>
        <StyledRow>
          <Flex alignItems="center" justifyContent="start">
            <IconBox>
              <Icons.Cluster />
            </IconBox>
            <H2>Cluster Information OG</H2>
          </Flex>
        </StyledRow>
        <StyledRow
          css={`
            padding-left: ${props => props.theme.space[6]}px;
          `}
        >
          <DataItem title="Cluster Name" data={clusterId} />
          <DataItem title="Teleport Version" data={authVersion} />
          <DataItem title="Public Address" data={publicURL} />
          {tunnelPublicAddress && (
            <DataItem title="Public SSH Tunnel" data={tunnelPublicAddress} />
          )}
          {isEnterprise && !cfg.isCloud && licenseExpiryDateText && (
            <DataItem title="License Expiry" data={licenseExpiryDateText} />
          )}
        </StyledRow>
      </StyledMultiRowBox>
      <MobileSeparator />

      <StyledMultiRowBox mb={3}>
        <StyledRow>
          <Flex alignItems="center" justifyContent="start">
            <IconBox>
              <Icons.Cluster />
            </IconBox>
            <H2>Services</H2>
          </Flex>
        </StyledRow>
        
        <StyledRow>
            <Text typography="body2" mb={2}>
            The following {serviceItems.length} services are running on {Array.from(new Set(serviceItems.map(item => item.Hostname))).length} unique hosts: 
            </Text>
        </StyledRow>
        <StyledRow
          css={`
            padding-left: ${props => props.theme.space[6]}px;
          `}
        >
          {/* To refresh the data after changes to serviceItems, manage serviceItems with useState and update it when data changes. */}


          <StyledTable

            columns={[
              { key: 'Hostname', headerText: 'Hostname' },
              { key: 'Name', headerText: 'Name' },
              { key: 'ProcessUptime', headerText: 'Process Uptime' },
              { key: 'Status', headerText: 'Status' },
              { key: 'Version', headerText: 'Version' },
            ]}
            data={serviceItems}
            rowKey="HostID-Name"
            renderRow={row => (
              <tr key={`${row.HostID}-${row.Name}`}>
          <Cell>{row.Hostname}</Cell>
          <Cell>{row.Name}</Cell>
          <Cell>{row.ProcessUptime}</Cell>
          <Cell>{row.Status}</Cell>
          <Cell>{row.Version}</Cell>
              </tr>
            )}
          />
        </StyledRow>

      </StyledMultiRowBox>
      <MobileSeparator />


      <StyledMultiRowBox mb={3}>
        <StyledRow>
          <SupportContentFlex
            alignItems="center"
            justifyContent="space-between"
          >
            <Flex alignItems="center">
              <IconBox>
                <Icons.Question />
              </IconBox>
              <H2>Support and Resource Pages</H2>
            </Flex>
            <SupportButtonBox>
              {showPremiumSupportCTA && (
                <ButtonLockedFeature event={CtaEvent.CTA_PREMIUM_SUPPORT}>
                  Unlock Premium Support w/Enterprise
                </ButtonLockedFeature>
              )}
            </SupportButtonBox>
          </SupportContentFlex>
        </StyledRow>
        <StyledRow
          css={`
            padding-left: ${props => props.theme.space[6]}px;
          `}
        >
          <SupportLinksFlex>
            <Box>
              <H3 ml={2} mb={1}>
                Support
              </H3>
              {isEnterprise && !showPremiumSupportCTA && (
                <ExternalSupportLink
                  title="Create a Support Ticket"
                  url="https://support.goteleport.com"
                />
              )}
              <ExternalSupportLink
                title="Ask the Community Questions"
                url="https://github.com/gravitational/teleport/discussions"
              />
              <ExternalSupportLink
                title="Request a New Feature"
                url="https://github.com/gravitational/teleport/issues/new/choose"
              />
              <ExternalSupportLink
                title="Send Product Feedback"
                url="mailto:support@goteleport.com"
              />
            </Box>
            <Box>
              <H3 ml={2} mb={1}>
                Resources
              </H3>
              <ExternalSupportLink
                title="Get Started Guide"
                url={docs.getStarted}
              />
              <ExternalSupportLink title="tsh User Guide" url={docs.tshGuide} />
              <ExternalSupportLink title="Admin Guides" url={docs.adminGuide} />
              <ExternalSupportLink
                title="Troubleshooting Guide"
                url={docs.troubleshooting}
              />
              <DownloadLink isCloud={isCloud} isEnterprise={isEnterprise} />
              <ExternalSupportLink title="FAQ" url={docs.faq} />
            </Box>
            <Box>
              <H3 ml={2} mb={1}>
                Updates
              </H3>
              <ExternalSupportLink
                title="Product Changelog"
                url={docs.changeLog}
              />
              <ExternalSupportLink
                title="Teleport Blog"
                url="https://goteleport.com/blog/"
              />
            </Box>
          </SupportLinksFlex>
        </StyledRow>
      </StyledMultiRowBox>
      {children}
    </FeatureBox>
  );




};

export const StyledMultiRowBox = styled(MultiRowBox)`
  @media screen and (max-width: ${props => props.theme.breakpoints.mobile}) {
    border: none;
  }
`;

export const StyledRow = styled(Row)`
  @media screen and (max-width: ${props => props.theme.breakpoints.mobile}) {
    border: none !important;
    padding-left: 0;
    padding-bottom: 0;
  }
`;

export const MobileSeparator = styled.div`
  width: 100vw;
  margin-left: -${props => props.theme.space[6]}px;
  @media screen and (max-width: ${props => props.theme.breakpoints.mobile}) {
    border-bottom: ${props =>
      `${props.theme.borders[1]} ${props.theme.colors.interactive.tonal.neutral[2]}`};
  }
`;

export const IconBox = styled(Box)`
  line-height: 0;
  padding: ${props => props.theme.space[2]}px;
  border-radius: ${props => props.theme.radii[3]}px;
  margin-right: ${props => props.theme.space[3]}px;
  background: ${props => props.theme.colors.interactive.tonal.neutral[0]};

  @media screen and (max-width: ${props => props.theme.breakpoints.mobile}) {
    background: transparent;
    margin-right: ${props => props.theme.space[1]}px;
  }
`;

const SupportContentFlex = styled(Flex)`
  @media screen and (max-width: ${props => props.theme.breakpoints.mobile}) {
    flex-direction: column;
    align-items: flex-start;
  }
`;

const SupportButtonBox = styled(Box)`
  width: 320px;
  @media screen and (max-width: ${props => props.theme.breakpoints.mobile}) {
    margin-left: ${props => props.theme.space[6]}px;
    margin-top: ${props => props.theme.space[2]}px;
  }
`;

const SupportLinksFlex = styled(Flex)`
  justify-content: space-between;
  flex-wrap: wrap;
  max-width: 70%;
  @media screen and (max-width: ${props => props.theme.breakpoints.tablet}) {
    max-width: 100%;
  }
  @media screen and (max-width: ${props => props.theme.breakpoints.mobile}) {
    flex-direction: column;
    gap: ${props => props.theme.space[3]}px;
    margin-bottom: ${props => props.theme.space[3]}px;
  }
`;

const DataItemFlex = styled(Flex)`
  margin-bottom: ${props => props.theme.space[3]}px;
  @media screen and (max-width: ${props => props.theme.breakpoints.mobile}) {
    flex-direction: column;
    padding-left: ${props => props.theme.space[2]}px;
  }
`;

/**
 * getDocUrls returns an object of URL's appended with
 * UTM, version, and type of teleport.
 *
 * @param version teleport version retrieved from cluster info.
 */
const getDocUrls = (version = '', isEnterprise: boolean) => {
  const verPrefix = isEnterprise ? 'e' : 'oss';

  /**
   * withUTM appends URL with UTM parameters.
   * anchor hashes must be appended at end of URL otherwise it is ignored.
   *
   * @param url the full link to the specific documentation.
   * @param anchorHash the hash in URL that predefines scroll location in the page.
   */
  const withUTM = (url = '', anchorHash = '') =>
    `${url}?product=teleport&version=${verPrefix}_${version}${anchorHash}`;

  return {
    getStarted: withUTM(`https://goteleport.com/docs/get-started/`),
    tshGuide: withUTM(`https://goteleport.com/docs/connect-your-client/tsh/`),
    adminGuide: withUTM(
      `https://goteleport.com/docs/admin-guides/management/admin/`
    ),
    faq: withUTM(`https://goteleport.com/docs/faq`),
    troubleshooting: withUTM(
      `https://goteleport.com/docs/admin-guides/management/admin/troubleshooting/`
    ),

    // there isn't a version-specific changelog page
    changeLog: withUTM('https://goteleport.com/docs/changelog'),
  };
};

const DownloadLink = ({
  isCloud,
  isEnterprise,
}: {
  isCloud: boolean;
  isEnterprise: boolean;
}) => {
  if (isCloud) {
    return (
      <StyledSupportLink as={Link} to={cfg.routes.downloadCenter}>
        Download Page
      </StyledSupportLink>
    );
  }

  if (isEnterprise) {
    return (
      <ExternalSupportLink
        title="Self-Hosting Teleport"
        url="https://goteleport.com/docs/admin-guides/deploy-a-cluster/"
      />
    );
  }

  return (
    <ExternalSupportLink
      title="Download Page"
      url="https://goteleport.com/download/"
    />
  );
};

const ExternalSupportLink = ({ title = '', url = '' }) => (
  <StyledSupportLink href={url} target="_blank">
    {title}
  </StyledSupportLink>
);

const StyledSupportLink = styled.a.attrs({
  rel: 'noreferrer',
})`
  display: block;
  color: ${props => props.theme.colors.text.main};
  border-radius: 4px;
  text-decoration: none;
  margin-bottom: 8px;
  padding: 4px 8px;
  transition: all 0.3s;

  ${props => props.theme.typography.body2}
  &:hover, &:focus {
    background: ${props => props.theme.colors.spotBackground[0]};
  }
`;


export const StyledTable = styled(Table)`
  td {
    height: 22px;
  }
` as typeof Table;


export const DataItem = ({ title = '', data = null }) => (
  <DataItemFlex>
    <Text typography="body2" bold style={{ width: '136px' }}>
      {title}:
    </Text>
    <Text typography="body2">{data}</Text>
  </DataItemFlex>
);

export type Props = {
  clusterId: string;
  authVersion: string;
  publicURL: string;
  licenseExpiryDateText?: string;
  isEnterprise: boolean;
  isCloud: boolean;
  tunnelPublicAddress?: string;
  children?: React.ReactNode;
  showPremiumSupportCTA: boolean;
  serviceItems?: any[];
};

