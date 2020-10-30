package usage

import (
	"context"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
	"gitlab.com/beneath-hq/beneath/models"
)

const (
	minReadBytesBilled  = 1024
	minWriteBytesBilled = 1024
	minScanBytesBilled  = 1048576
)

// TrackRead is a helper to track read usage for a secret from an instance
func (s *Service) TrackRead(ctx context.Context, secret models.Secret, streamID uuid.UUID, instanceID uuid.UUID, nrecords int64, nbytes int64) {
	if streamID != uuid.Nil {
		s.trackOp(OpTypeRead, streamID, time.Time{}, nrecords, nbytes)
	}
	if instanceID != uuid.Nil {
		s.trackOp(OpTypeRead, instanceID, time.Time{}, nrecords, nbytes)
	}
	if !secret.IsAnonymous() {
		if nbytes < minReadBytesBilled {
			nbytes = minReadBytesBilled
		}
		s.trackOp(OpTypeRead, secret.GetOwnerID(), secret.GetOwnerQuotaEpoch(), nrecords, nbytes)
		s.trackOp(OpTypeRead, secret.GetBillingOrganizationID(), secret.GetBillingQuotaEpoch(), nrecords, nbytes)
	}
}

// TrackWrite is a helper to track write usage for a secret to an instance
func (s *Service) TrackWrite(ctx context.Context, secret models.Secret, streamID uuid.UUID, instanceID uuid.UUID, nrecords int64, nbytes int64) {
	s.trackOp(OpTypeWrite, streamID, time.Time{}, nrecords, nbytes)
	s.trackOp(OpTypeWrite, instanceID, time.Time{}, nrecords, nbytes)
	if !secret.IsAnonymous() {
		if nbytes < minWriteBytesBilled {
			nbytes = minWriteBytesBilled
		}
		s.trackOp(OpTypeWrite, secret.GetOwnerID(), secret.GetOwnerQuotaEpoch(), nrecords, nbytes)
		s.trackOp(OpTypeWrite, secret.GetBillingOrganizationID(), secret.GetBillingQuotaEpoch(), nrecords, nbytes)
	}
}

// TrackScan is a helper to track scan usage for a secret to an instance
func (s *Service) TrackScan(ctx context.Context, secret models.Secret, nbytes int64) {
	if !secret.IsAnonymous() {
		if nbytes < minScanBytesBilled {
			nbytes = minScanBytesBilled
		}
		s.trackOp(OpTypeScan, secret.GetOwnerID(), secret.GetOwnerQuotaEpoch(), 0, nbytes)
		s.trackOp(OpTypeScan, secret.GetBillingOrganizationID(), secret.GetBillingQuotaEpoch(), 0, nbytes)
	}
}

// trackOp records the usage for a given ID (stream, instance, user, service, organization, ...)
func (s *Service) trackOp(op OpType, id uuid.UUID, maybeQuotaEpoch time.Time, nrecords int64, nbytes int64) {
	s.mu.Lock()
	u := s.usageBuffer[id]
	if op == OpTypeRead {
		u.ReadOps++
		u.ReadRecords += nrecords
		u.ReadBytes += nbytes
	} else if op == OpTypeWrite {
		u.WriteOps++
		u.WriteRecords += nrecords
		u.WriteBytes += nbytes
	} else if op == OpTypeScan {
		u.ScanOps++
		u.ScanBytes += nbytes
	} else {
		s.mu.Unlock()
		panic(fmt.Errorf("unrecognized op type '%d'", op))
	}
	s.usageBuffer[id] = u
	if !maybeQuotaEpoch.IsZero() {
		s.quotaEpochBuffer[id] = maybeQuotaEpoch
	}
	s.mu.Unlock()
}
