package role

import "fmt"

type RoleManager struct {
	roleLevelMap map[string]int
	topRole      string
	bottomRole   string
}

// NewRoleManager creates a new RoleManager with the provided roles. The first
// role in the list is considered the highest level, and the last role is
// considered the lowest level.
func NewRoleManager(roles ...string) *RoleManager {
	rm := &RoleManager{
		roleLevelMap: make(map[string]int),
	}

	len := len(roles)
	for i, role := range roles {
		rm.roleLevelMap[role] = len - i
	}

	rm.topRole = roles[0]
	rm.bottomRole = roles[len-1]

	return rm
}

// HasMultipleRoles checks if the RoleManager has more than one role.
func (rm *RoleManager) HasMultipleRoles() bool {
	return len(rm.roleLevelMap) > 1
}

// GetTopRole returns the role with the highest level in the RoleManager.
func (rm *RoleManager) GetTopRole() string {
	return rm.topRole
}

// GetBottomRole returns the role with the lowest level in the RoleManager.
func (rm *RoleManager) GetBottomRole() string {
	return rm.bottomRole
}

// HasRole checks if a role exists in the RoleManager.
func (rm *RoleManager) HasRole(role string) bool {
	_, exists := rm.roleLevelMap[role]
	return exists
}

// CompareRoles compares the levels of two roles. It returns a positive integer
// if role1 is higher than role2, a negative integer if role1 is lower than
// role2, and zero if they are equal. If either role does not exist, it returns
// an error.
func (rm *RoleManager) CompareRoles(role1, role2 string) (int, error) {
	if !rm.HasRole(role1) || !rm.HasRole(role2) {
		return 0, fmt.Errorf("one or both roles not found")
	}
	return rm.roleLevelMap[role1] - rm.roleLevelMap[role2], nil
}
