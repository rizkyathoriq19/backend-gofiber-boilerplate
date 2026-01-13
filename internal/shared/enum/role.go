package enum

type UserRole string

const (
	UserRoleSuperAdmin UserRole = "super_admin"
	UserRoleManager    UserRole = "manager"
	UserRoleHeadNurse  UserRole = "head_nurse"
	UserRoleNurse      UserRole = "nurse"
	UserRoleDoctor     UserRole = "doctor"
	UserRolePatient    UserRole = "patient"
)

func (r UserRole) String() string {
	return string(r)
}

func (r UserRole) IsValid() bool {
	switch r {
	case UserRoleSuperAdmin, UserRoleManager, UserRoleHeadNurse, UserRoleNurse, UserRoleDoctor, UserRolePatient:
		return true
	default:
		return false
	}
}

// IsSuperAdmin checks if role has super admin privileges
func (r UserRole) IsSuperAdmin() bool {
	return r == UserRoleSuperAdmin
}

// IsAdmin checks if role has admin privileges (super_admin or manager)
func (r UserRole) IsAdmin() bool {
	return r == UserRoleSuperAdmin || r == UserRoleManager
}

// IsStaff checks if role is a staff member (not patient)
func (r UserRole) IsStaff() bool {
	switch r {
	case UserRoleSuperAdmin, UserRoleManager, UserRoleHeadNurse, UserRoleNurse, UserRoleDoctor:
		return true
	default:
		return false
	}
}

// CanViewAllRooms checks if role can view all rooms
func (r UserRole) CanViewAllRooms() bool {
	switch r {
	case UserRoleSuperAdmin, UserRoleManager, UserRoleHeadNurse, UserRoleDoctor:
		return true
	default:
		return false
	}
}
