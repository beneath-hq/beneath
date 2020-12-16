import { FC } from "react";

import { EntityKind } from "apollo/types/globalTypes";
import { ServiceByOrganizationProjectAndName_serviceByOrganizationProjectAndName } from "apollo/types/ServiceByOrganizationProjectAndName";
import OwnerUsageView from "components/usage/OwnerUsageView";

export interface ViewUsageProps {
  service: ServiceByOrganizationProjectAndName_serviceByOrganizationProjectAndName;
}

const ViewUsage: FC<ViewUsageProps> = ({ service }) => {
  return (
    <OwnerUsageView
      entityKind={EntityKind.Service}
      entityID={service.serviceID}
      readQuota={service.readQuota}
      writeQuota={service.writeQuota}
      scanQuota={service.scanQuota}
      quotaStartTime={service.quotaStartTime}
      quotaEndTime={service.quotaEndTime}
    />
  );
};

export default ViewUsage;
