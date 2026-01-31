package keycloak

import (
	"context"
	"fmt"
	"log"
)

var Roles = []struct {
	Name string
	Desc string
}{
	{"super_admin", "Super Administrator with full system access"},
	{"admin", "Administrator with management access"},
	{"manager", "Department manager with oversight access"},
	{"data_entry", "Data entry personnel"},
	{"viewer", "Read-only access"},
	{"lab_technician", "Laboratory technician"},
	{"surveillance_officer", "Surveillance officer"},
}

type Action string

const (
	Create Action = "create"
	Read   Action = "read"
	Update Action = "update"
	Delete Action = "delete"
	Export Action = "export"
)

type Permission struct {
	Resource string
	Actions  []Action
}

var Permissions = []Permission{
	{Resource: "users", Actions: []Action{Create, Read, Update, Delete}},
	{Resource: "vhf_patients", Actions: []Action{Create, Read, Update, Delete, Export}},
	{Resource: "outbreaks", Actions: []Action{Create, Read, Update, Delete}},
	{Resource: "reports", Actions: []Action{Read, Export}},
	{Resource: "employees", Actions: []Action{Create, Read, Update, Delete}},
	{Resource: "facilities", Actions: []Action{Create, Read, Update, Delete}},
	{Resource: "laboratory", Actions: []Action{Create, Read, Update, Delete}},
	{Resource: "surveillance", Actions: []Action{Create, Read, Update, Delete}},
}

var RolePresets = map[string]func(resource string, action Action) bool{
	"super_admin": func(_ string, _ Action) bool {
		return true
	},

	"admin": func(resource string, _ Action) bool {
		return resource != "users"
	},

	"manager": func(_ string, action Action) bool {
		return action == Read || action == Update || action == Export
	},

	"data_entry": func(_ string, action Action) bool {
		return action == Create || action == Read
	},

	"viewer": func(_ string, action Action) bool {
		return action == Read
	},

	"lab_technician": func(resource string, _ Action) bool {
		return resource == "laboratory"
	},

	"surveillance_officer": func(resource string, _ Action) bool {
		return resource == "surveillance"
	},
}

func BootstrapRBAC(ctx context.Context, kc *AdminClient, clientID string) error {
	log.Println("🔐 Bootstrapping RBAC into Keycloak...")

	// -------------------------
	// 1. Realm Roles
	// -------------------------
	for _, r := range Roles {
		if err := kc.CreateRealmRoleIfMissing(ctx, r.Name, r.Desc); err != nil {
			return err
		}
	}

	// -------------------------
	// 2. Scopes
	// -------------------------
	scopeSet := map[string]struct{}{}
	for _, p := range Permissions {
		for _, a := range p.Actions {
			scopeSet[string(a)] = struct{}{}
		}
	}

	for scope := range scopeSet {
		if err := kc.CreateScopeIfMissing(ctx, clientID, scope); err != nil {
			return err
		}
	}

	// -------------------------
	// 3. Resources
	// -------------------------
	for _, p := range Permissions {
		if err := kc.CreateResourceIfMissing(ctx, clientID, p.Resource); err != nil {
			return err
		}
	}

	// -------------------------
	// 4. Permissions
	// -------------------------
	for _, p := range Permissions {
		scopes := make([]string, 0, len(p.Actions))
		for _, a := range p.Actions {
			scopes = append(scopes, string(a))
		}

		if err := kc.CreatePermissionIfMissing(
			ctx,
			clientID,
			fmt.Sprintf("%s-permission", p.Resource),
			p.Resource,
			scopes,
		); err != nil {
			return err
		}
	}

	// -------------------------
	// 5. Role Policies + Attach
	// -------------------------
	for role, rule := range RolePresets {
		policyName := role + "-policy"

		if err := kc.CreateRolePolicyIfMissing(ctx, clientID, policyName, role); err != nil {
			return err
		}

		for _, p := range Permissions {
			allowedScopes := []string{}

			for _, a := range p.Actions {
				if rule(p.Resource, a) {
					allowedScopes = append(allowedScopes, string(a))
				}
			}

			if len(allowedScopes) == 0 {
				continue
			}

			if err := kc.AttachPolicyToPermission(
				ctx,
				clientID,
				fmt.Sprintf("%s-permission", p.Resource),
				policyName,
			); err != nil {
				return err
			}
		}
	}

	log.Println("✅ RBAC bootstrap complete")
	return nil
}

func EnsureDefaultAdmin(ctx context.Context, kc *AdminClient) error {
	userID, err := kc.EnsureUser(
		ctx,
		"admin",
		"admin@system.local",
		"System",
		"Administrator",
		true,
	)
	if err != nil {
		return err
	}

	return kc.AssignRealmRole(ctx, userID, "super_admin")
}
