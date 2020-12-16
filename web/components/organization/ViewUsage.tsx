import { FC } from "react";

import { EntityKind } from "apollo/types/globalTypes";
import { OrganizationByName_organizationByName_PrivateOrganization} from "apollo/types/OrganizationByName";
import OwnerUsageView, { OwnerUsageViewProps } from "components/usage/OwnerUsageView";

export interface ViewUsageProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
}

const ViewUsage: FC<ViewUsageProps> = ({ organization }) => {
  // show metrics on org level, unless it's a personal org that doesn't handle billing (then show on user-level)
  let ownerUsageViewProps: OwnerUsageViewProps;
  if (organization.personalUser && organization.personalUser.billingOrganizationID !== organization.organizationID) {
    ownerUsageViewProps = {
      entityKind: EntityKind.User,
      entityID: organization.personalUser.userID,
      readQuota: organization.personalUser.readQuota,
      writeQuota: organization.personalUser.writeQuota,
      scanQuota: organization.personalUser.scanQuota,
      quotaStartTime: organization.personalUser.quotaStartTime,
      quotaEndTime: organization.personalUser.quotaEndTime,
    };
  } else {
    ownerUsageViewProps = {
      entityKind: EntityKind.Organization,
      entityID: organization.organizationID,
      readQuota: organization.prepaidReadQuota,
      writeQuota: organization.prepaidWriteQuota,
      scanQuota: organization.prepaidScanQuota,
      prepaidReadQuota: organization.prepaidReadQuota,
      prepaidWriteQuota: organization.prepaidWriteQuota,
      prepaidScanQuota: organization.prepaidScanQuota,
      quotaStartTime: organization.quotaStartTime,
      quotaEndTime: organization.quotaEndTime,
    };
  }

  return <OwnerUsageView {...ownerUsageViewProps} />
};

export default ViewUsage;
