package prooftransport

import (
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"time"
)

func (w *wire) observe(ctx context.Context, material []byte) R_TransportObservation {
	out := R_TransportObservation{F_base_url: w.scope.F_base_url, F_material_digest: digest(material), F_observed_at: time.Now().Unix()}
	for _, item := range []struct {
		path   string
		status *int64
		digest *string
	}{
		{"/manifests/" + w.scope.F_subject_digest, &out.F_subject_status, &out.F_subject_digest},
		{"/manifests/" + w.scope.F_manifest_digest, &out.F_manifest_status, &out.F_manifest_digest},
		{"/blobs/" + w.scope.F_payload_digest, &out.F_payload_status, &out.F_payload_digest},
		{"/blobs/" + w.scope.F_config_digest, &out.F_config_status, &out.F_config_digest},
	} {
		res, err := w.read(ctx, w.scope.F_base_url+item.path)
		*item.status, *item.digest = res.status, digest(res.data)
		if err != nil {
			out.F_read_error = err.Error()
			out.F_now = time.Now().Unix()
			return out
		}
	}
	if err := w.referrers(ctx, &out); err != nil {
		out.F_read_error = err.Error()
	}
	out.F_now = time.Now().Unix()
	return out
}

func (w *wire) referrers(ctx context.Context, out *R_TransportObservation) error {
	out.F_referrers_url = w.scope.F_referrers_url
	res, err := w.read(ctx, out.F_referrers_url)
	if err == nil && res.status == 404 {
		out.F_referrers_url = w.scope.F_fallback_url
		res, err = w.read(ctx, out.F_referrers_url)
	}
	out.F_referrers_status = res.status
	if err != nil {
		return err
	}
	out.F_referrers_digest = digest(res.data)
	out.F_next_link = res.header.Get("Link")
	out.F_referrers_media_type, _, _ = mime.ParseMediaType(res.header.Get("Content-Type"))
	if res.status != 200 {
		return nil
	}
	var index struct {
		SchemaVersion int    `json:"schemaVersion"`
		MediaType     string `json:"mediaType"`
		Manifests     *[]struct {
			Digest string `json:"digest"`
		} `json:"manifests"`
	}
	if err := json.Unmarshal(res.data, &index); err != nil {
		return fmt.Errorf("invalid referrers JSON")
	}
	if index.SchemaVersion != 2 || index.MediaType != out.F_referrers_media_type || index.Manifests == nil {
		return fmt.Errorf("incomplete referrers index")
	}
	for _, descriptor := range *index.Manifests {
		out.F_descriptors = append(out.F_descriptors, descriptor.Digest)
	}
	return nil
}
