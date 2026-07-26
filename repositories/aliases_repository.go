package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/alias"
	"arturgudiev/dashboard/ent/predicate"
	"arturgudiev/dashboard/ent/schema"
	"context"
	"path/filepath"
	"strings"
)

type AliasesRepository struct {
	client *ent.Client
}

func NewAliasesRepository(client *ent.Client) *AliasesRepository {
	return &AliasesRepository{client: client}
}

func (r *AliasesRepository) GetAliasByAliasString(ctx context.Context, aliasString string) (*ent.Alias, error) {
	aliasEntity, err := r.client.Alias.Query().Where(alias.AliasEQ(aliasString)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return aliasEntity, nil
}

func (r *AliasesRepository) GetAliasesByAliasType(
	ctx context.Context,
	aliasType schema.AliasType,
	taskContainerID int,
) ([]*ent.Alias, error) {
	aliases, err := r.client.Alias.Query().Where(
		alias.ItemIDEQ(taskContainerID),
		alias.TypeEQ(aliasType),
	).All(ctx)
	if err != nil {
		return nil, err
	}
	return aliases, nil
}

func (r *AliasesRepository) GetAliasesByFilePath(
	ctx context.Context,
	filePath string,
) ([]*ent.Alias, error) {
	clean := filepath.Clean(strings.TrimSpace(filePath))
	if clean == "" || clean == "." {
		return nil, nil
	}
	forward := strings.ReplaceAll(clean, `\`, `/`)
	back := strings.ReplaceAll(clean, `/`, `\`)
	forwardNoBin := strings.TrimSuffix(forward, ".bin")
	backNoBin := strings.TrimSuffix(back, ".bin")

	seen := make(map[string]struct{})
	paths := make([]string, 0, 8)
	for _, p := range []string{clean, forward, back, forwardNoBin, backNoBin, forwardNoBin + ".bin", backNoBin + ".bin"} {
		if p == "" || p == ".bin" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		paths = append(paths, p)
	}

	preds := make([]predicate.Alias, len(paths))
	for i, p := range paths {
		// Case-insensitive (and DB collation-safe) match: paths from the app may
		// differ in letter case from stored aliases, e.g. ...\dashboard\... vs ...\Dashboard\...
		preds[i] = alias.FilePathEqualFold(p)
	}

	// Also pull likely candidates by relative suffix (Windows abs vs local FILES_DIR abs).
	rel := logicalFileRelPath(clean)
	if rel != "" {
		relForward := strings.ReplaceAll(rel, `\`, `/`)
		relBack := strings.ReplaceAll(rel, `/`, `\`)
		for _, suffix := range []string{relForward, relBack, relForward + ".bin", relBack + ".bin"} {
			preds = append(preds, alias.FilePathHasSuffix(suffix))
			preds = append(preds, alias.FilePathContainsFold(suffix))
		}
	}

	var match predicate.Alias
	switch len(preds) {
	case 0:
		return nil, nil
	case 1:
		match = preds[0]
	default:
		match = alias.Or(preds...)
	}

	candidates, err := r.client.Alias.Query().Where(
		alias.TypeEQ(schema.AliasTypeFile),
		alias.FilePathNotNil(),
		match,
	).All(ctx)
	if err != nil {
		return nil, err
	}

	// Exact EqualFold hits are kept; Contains/HasSuffix hits are filtered with logical matching
	// so different absolute roots (e.g. Windows vs macOS FILES_DIR) still resolve.
	out := make([]*ent.Alias, 0, len(candidates))
	seenIDs := make(map[int]struct{}, len(candidates))
	for _, a := range candidates {
		if a.FilePath == nil {
			continue
		}
		if _, dup := seenIDs[a.ID]; dup {
			continue
		}
		if fileAliasPathMatches(*a.FilePath, clean) || (rel != "" && fileAliasPathMatches(*a.FilePath, rel)) {
			seenIDs[a.ID] = struct{}{}
			out = append(out, a)
		}
	}
	return out, nil
}

func (r *AliasesRepository) CreateFileAlias(ctx context.Context, alias string, filePath string) (*ent.Alias, error) {
	aliasEntity, err := r.client.Alias.Create().SetFilePath(filePath).SetAlias(alias).SetType(schema.AliasTypeFile).Save(ctx)
	if err != nil {
		return nil, err
	}
	return aliasEntity, nil
}

func (r *AliasesRepository) RemoveFileAlias(ctx context.Context, aliasStr string, filePath string) (*ent.Alias, error) {
	aliasEntity, err := r.client.Alias.Query().Where(
		alias.FilePathEQ(filePath),
		alias.AliasEQ(aliasStr),
	).Only(ctx)
	if err != nil {
		return nil, err
	}
	if err := r.client.Alias.DeleteOneID(aliasEntity.ID).Exec(ctx); err != nil {
		return nil, err
	}
	return aliasEntity, nil
}

func (r *AliasesRepository) CreateAliasByContainerType(ctx context.Context, aliasType schema.AliasType, itemID int, alias string) (*ent.Alias, error) {
	aliasEntity, err := r.client.Alias.Create().SetType(aliasType).SetItemID(itemID).SetAlias(alias).Save(ctx)
	if err != nil {
		return nil, err
	}
	return aliasEntity, nil
}

func (r *AliasesRepository) RemoveAliasFromContainer(ctx context.Context, aliasType schema.AliasType, itemID int, aliasStr string) (*ent.Alias, error) {
	aliasEntity, err := r.client.Alias.Query().Where(
		alias.TypeEQ(aliasType),
		alias.ItemIDEQ(itemID),
		alias.AliasEQ(aliasStr),
	).Only(ctx)
	if err != nil {
		return nil, err
	}
	if err := r.client.Alias.DeleteOneID(aliasEntity.ID).Exec(ctx); err != nil {
		return nil, err
	}
	return aliasEntity, nil
}
