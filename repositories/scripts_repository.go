package repositories

import (
	"arturgudiev/dashboard/ent"
	"arturgudiev/dashboard/ent/schema"
	"arturgudiev/dashboard/ent/script"
	"context"
	"strings"
)

type ScriptsRepository struct {
	client *ent.Client
}

func NewScriptsRepository(client *ent.Client) *ScriptsRepository {
	return &ScriptsRepository{client: client}
}

func (r *ScriptsRepository) Create(
	ctx context.Context,
	name, code, description string,
	params []schema.ScriptParam,
	containerType *schema.ContainerType,
	containerID *int,
) (*ent.Script, error) {
	if params == nil {
		params = []schema.ScriptParam{}
	}
	builder := r.client.Script.Create().
		SetName(name).
		SetCode(code).
		SetDescription(description).
		SetParams(params)
	if containerType != nil {
		builder = builder.SetContainerType(*containerType)
	}
	if containerID != nil {
		builder = builder.SetContainerID(*containerID)
	}
	return builder.Save(ctx)
}

func (r *ScriptsRepository) Get(ctx context.Context, id int) (*ent.Script, error) {
	return r.client.Script.Get(ctx, id)
}

func (r *ScriptsRepository) List(
	ctx context.Context,
	query string,
	scope string,
	containerType schema.ContainerType,
	containerID int,
) ([]*ent.Script, error) {
	q := r.client.Script.Query().Order(ent.Asc(script.FieldName))
	if trimmed := strings.TrimSpace(query); trimmed != "" {
		q = q.Where(script.NameContainsFold(trimmed))
	}

	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "global":
		q = q.Where(script.ContainerIDIsNil())
	case "local":
		q = q.Where(
			script.ContainerTypeEQ(containerType),
			script.ContainerIDEQ(containerID),
		)
	case "all":
		fallthrough
	default:
		if containerID > 0 && containerType != "" {
			q = q.Where(
				script.Or(
					script.ContainerIDIsNil(),
					script.And(
						script.ContainerTypeEQ(containerType),
						script.ContainerIDEQ(containerID),
					),
				),
			)
		} else {
			q = q.Where(script.ContainerIDIsNil())
		}
	}

	return q.All(ctx)
}

func (r *ScriptsRepository) Update(
	ctx context.Context,
	id int,
	name, code, description *string,
	params *[]schema.ScriptParam,
) (*ent.Script, error) {
	upd := r.client.Script.UpdateOneID(id)
	if name != nil {
		upd = upd.SetName(*name)
	}
	if code != nil {
		upd = upd.SetCode(*code)
	}
	if description != nil {
		upd = upd.SetDescription(*description)
	}
	if params != nil {
		upd = upd.SetParams(*params)
	}
	return upd.Save(ctx)
}

func (r *ScriptsRepository) Delete(ctx context.Context, id int) error {
	return r.client.Script.DeleteOneID(id).Exec(ctx)
}
