package lsp

import "context"

func (server *Server) ObserveRefreshStageProvenancePart01(
    ctx context.Context,
    uri string,
) (RefreshStageProvenancePart01, error) {
    observation, err := server.ObserveRefreshPart01(ctx, uri)
    if err != nil {
        return RefreshStageProvenancePart01{}, err
    }
    return BuildRefreshStageProvenancePart01(observation)
}
