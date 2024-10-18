package huaweicloud

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
	"github.com/libdns/libdns"
)

// Provider facilitates DNS record manipulation with Huawei Cloud
type Provider struct {
	// AccessKeyId is required by the Huawei Cloud API for authentication.
	AccessKeyId string `json:"access_key_id,omitempty"`
	// SecretAccessKey is required by the Huawei Cloud API for authentication.
	SecretAccessKey string `json:"secret_access_key,omitempty"`
	// RegionId is optional and defaults to "cn-south-1".
	RegionId string `json:"region_id,omitempty"`
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	zoneId, err := p.getZoneIdByName(zone)
	fmt.Println(zoneId)
	fmt.Println(err)
	if err != nil {
		return nil, err
	}

	client, err := p.getClient()
	if err != nil {
		return nil, err
	}

	request := &model.ListRecordSetsByZoneRequest{
		ZoneId: zoneId,
	}
	response, err := client.ListRecordSetsByZone(request)
	if err != nil {
		return nil, err
	}

	var list []libdns.Record
	for record := range slices.Values(*response.Recordsets) {
		for value := range slices.Values(*record.Records) {
			list = append(list, libdns.Record{
				ID:    *record.Id,
				Type:  *record.Type,
				Name:  libdns.RelativeName(*record.Name, zone),
				Value: value,
				TTL:   time.Duration(*record.Ttl) * time.Second,
			})
		}
	}

	return list, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zoneId, err := p.getZoneIdByName(zone)
	if err != nil {
		return nil, err
	}
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}

	for record := range slices.Values(records) {
		ttl := int32(record.TTL.Seconds())
		request := &model.CreateRecordSetRequest{
			ZoneId: zoneId,
			Body: &model.CreateRecordSetRequestBody{
				Name:    libdns.AbsoluteName(record.Name, zone),
				Type:    record.Type,
				Ttl:     &ttl,
				Records: []string{record.Value},
			},
		}
		if _, err = client.CreateRecordSet(request); err != nil {
			return nil, err
		}
	}

	return records, nil
}

// SetRecords sets the records in the zone, either by updating existing records or creating new ones.
// It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zoneId, err := p.getZoneIdByName(zone)
	if err != nil {
		return nil, err
	}
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}

	for record := range slices.Values(records) {
		if record.ID == "" {
			if _, err = p.AppendRecords(ctx, zone, []libdns.Record{record}); err != nil {
				return nil, err
			}
		} else {
			name := libdns.AbsoluteName(record.Name, zone)
			ttl := int32(record.TTL.Seconds())
			value := []string{record.Value}
			request := &model.UpdateRecordSetRequest{
				ZoneId:      zoneId,
				RecordsetId: record.ID,
				Body: &model.UpdateRecordSetReq{
					Name:    &name,
					Type:    &record.Type,
					Ttl:     &ttl,
					Records: &value,
				},
			}
			if _, err = client.UpdateRecordSet(request); err != nil {
				return nil, err
			}
		}
	}

	return records, nil
}

// DeleteRecords deletes the records from the zone. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	zoneId, err := p.getZoneIdByName(zone)
	if err != nil {
		return nil, err
	}
	client, err := p.getClient()
	if err != nil {
		return nil, err
	}

	for record := range slices.Values(records) {
		request := &model.DeleteRecordSetRequest{
			ZoneId:      zoneId,
			RecordsetId: record.ID,
		}
		if _, err = client.DeleteRecordSet(request); err != nil {
			return nil, err
		}
	}

	return records, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
