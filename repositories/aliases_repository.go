package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/alias"
	"context"
)

type AliasesRepository struct {
	client *ent.Client
}

func NewAliasesRepository(client *ent.Client) *AliasesRepository {
	return &AliasesRepository{client: client}
}

func (r *AliasesRepository) GetAliasByAliasString(ctx context.Context, aliasString string) ([]*ent.Alias, error) {
	alias, err := r.client.Alias.Query().Where(alias.Alias(aliasString)).All(ctx)
	if err != nil {
		return nil, err
	}
	return alias, nil
}
