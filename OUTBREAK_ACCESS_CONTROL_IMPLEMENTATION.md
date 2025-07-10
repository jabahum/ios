# Outbreak Access Control Implementation Guide

## Overview

This implementation provides comprehensive access control for different outbreak types (VHF, MPOX) with role-based permissions and outbreak-specific user assignments.

## Features Implemented

### 1. Outbreak Type Classification
- **VHF Outbreaks**: Viral Hemorrhagic Fever outbreaks (Ebola, etc.)
- **MPOX Outbreaks**: Monkeypox outbreaks
- **General Outbreaks**: Other types of outbreaks

### 2. Role-Based Access Control
- **VHF Lab Technician**: Can access VHF list and capture lab requests
- **VHF Data Entry**: Can access VHF cases and enter data
- **MPOX Case Manager**: Can manage MPOX outbreak cases
- **MPOX Data Entry**: Can enter MPOX case data
- **Outbreak Viewer**: Can view assigned outbreaks but not edit/close
- **Outbreak Manager**: Can manage assigned outbreaks including edit/close

### 3. Outbreak Assignment System
- Users can be assigned to specific outbreaks
- Users assigned to outbreaks cannot see edit/close buttons unless they have manager permissions
- Role-based access automatically grants access to outbreak types

## Database Changes

### Migration: `migrations/017_add_outbreak_type_and_access_control.sql`

This migration adds:
- `outbreak_type` and `outbreak_category` columns to outbreaks table
- New roles for VHF and MPOX specific access
- New permissions for outbreak-specific operations
- Database functions for access control

### Key Functions:
- `check_user_outbreak_access(user_id, outbreak_id)`: Checks if user has access to outbreak
- `get_user_accessible_outbreaks(user_id)`: Gets outbreaks accessible to user
- `can_user_manage_outbreak(user_id, outbreak_id)`: Checks if user can edit/close outbreak

## User Setup

### Running the Setup Script

1. **Apply the migration first:**
```sql
-- Run the migration file
\i migrations/017_add_outbreak_type_and_access_control.sql
```

2. **Create test users with roles:**
```bash
go run scripts/setup_outbreak_users.go
```

### Default User Credentials

| Username | Password | Role | Access |
|----------|----------|------|--------|
| vhf_lab | password123 | VHF Lab Technician | VHF outbreaks, lab requests |
| vhf_data | password123 | VHF Data Entry | VHF outbreaks, data entry |
| mpox_manager | password123 | MPOX Case Manager | MPOX outbreaks, case management |
| mpox_data | password123 | MPOX Data Entry | MPOX outbreaks, data entry |
| outbreak_viewer | password123 | Outbreak Viewer | View assigned outbreaks only |
| outbreak_manager | password123 | Outbreak Manager | Manage assigned outbreaks |

## Access Control Rules

### 1. VHF Users
- **VHF Lab Technician** and **VHF Data Entry** roles automatically get access to all VHF outbreaks
- Can access `/vhf-list` and lab request functionality
- Cannot see MPOX outbreaks

### 2. MPOX Users
- **MPOX Case Manager** and **MPOX Data Entry** roles automatically get access to all MPOX outbreaks
- Can access MPOX case management and follow-up
- Cannot see VHF outbreaks

### 3. Outbreak Assignment
- Users assigned to specific outbreaks can only see those outbreaks
- Users with **outbreak_viewer** role can view but not edit/close
- Users with **outbreak_manager** role can edit/close assigned outbreaks
- **Super Admin** and **Admin** roles can manage all outbreaks

### 4. Edit/Close Permissions
- Only users with **outbreak_manager** role or higher can see edit/close buttons
- Users assigned to outbreaks without manager role cannot edit/close
- Admin roles can always edit/close any outbreak

## Implementation Details

### Updated Files

1. **Database Schema**: `migrations/017_add_outbreak_type_and_access_control.sql`
2. **Outbreak Model**: `internal/models/outbreak.xo.go` (added new fields and methods)
3. **Outbreak Handlers**: `internal/handlers/outbreaks.go` (added access control)
4. **Templates**: 
   - `ui/html/outbreaks.html` (shows outbreak type, conditional edit/close)
   - `ui/html/form_outbreak.html` (added outbreak type field)

### Key Methods Added

```go
// Check if user has access to outbreak
hasAccess, err := models.CheckUserOutbreakAccess(ctx, db, userID, outbreakID)

// Get outbreaks accessible to user
outbreaks, err := models.GetUserAccessibleOutbreaks(ctx, db, userID)

// Check if user can manage outbreak (edit/close)
canManage, err := models.CanUserManageOutbreak(ctx, db, userID, outbreakID)
```

## Usage Examples

### 1. VHF Lab Technician Login
- User logs in with `vhf_lab` / `password123`
- Sees only VHF outbreaks in outbreak list
- Can access VHF case list and lab requests
- Cannot see MPOX outbreaks or edit/close buttons

### 2. MPOX Case Manager Login
- User logs in with `mpox_manager` / `password123`
- Sees only MPOX outbreaks in outbreak list
- Can access MPOX case management
- Cannot see VHF outbreaks

### 3. Outbreak Manager Login
- User logs in with `outbreak_manager` / `password123`
- Sees outbreaks they're assigned to
- Can edit/close assigned outbreaks
- Cannot edit/close outbreaks they're not assigned to

## Testing the Implementation

### 1. Test VHF Access
```bash
# Login as VHF Lab Technician
curl -X POST http://localhost:3000/login \
  -d "username=vhf_lab&password=password123"

# Should only see VHF outbreaks
curl http://localhost:3000/outbreaks
```

### 2. Test MPOX Access
```bash
# Login as MPOX Case Manager
curl -X POST http://localhost:3000/login \
  -d "username=mpox_manager&password=password123"

# Should only see MPOX outbreaks
curl http://localhost:3000/outbreaks
```

### 3. Test Outbreak Assignment
```bash
# Login as Outbreak Viewer
curl -X POST http://localhost:3000/login \
  -d "username=outbreak_viewer&password=password123"

# Should see assigned outbreaks only
curl http://localhost:3000/outbreaks
```

## Security Considerations

1. **Role-Based Access**: Users can only access outbreaks based on their roles
2. **Assignment Control**: Users assigned to outbreaks cannot edit/close unless they have manager role
3. **Type Isolation**: VHF users cannot see MPOX outbreaks and vice versa
4. **Permission Inheritance**: Admin roles override all restrictions

## Troubleshooting

### Common Issues

1. **User cannot see any outbreaks**
   - Check if user has appropriate role assigned
   - Verify user is assigned to outbreaks
   - Check database functions are working

2. **Edit/Close buttons not showing**
   - Verify user has `outbreak_manager` role or higher
   - Check if user is assigned to the outbreak
   - Ensure `can_user_manage_outbreak` function is working

3. **Wrong outbreak types showing**
   - Verify outbreak has correct `outbreak_type` set
   - Check user role matches outbreak type
   - Ensure database functions are filtering correctly

### Debug Queries

```sql
-- Check user roles
SELECT u.user_name, r.name as role_name 
FROM users u 
JOIN user_roles ur ON u.user_id = ur.user_id 
JOIN roles r ON ur.role_id = r.id 
WHERE u.user_name = 'vhf_lab';

-- Check outbreak assignments
SELECT u.user_name, o.name as outbreak_name, o.outbreak_type
FROM users u 
JOIN user_outbreaks uo ON u.user_id = uo.user_id 
JOIN outbreaks o ON uo.outbreak_id = o.id 
WHERE uo.is_active = true;

-- Test access control function
SELECT check_user_outbreak_access(1, 1);
SELECT get_user_accessible_outbreaks(1);
SELECT can_user_manage_outbreak(1, 1);
```

## Next Steps

1. **Apply the migration** to add outbreak types and access control
2. **Run the setup script** to create test users
3. **Test the access control** with different user roles
4. **Customize roles and permissions** as needed for your organization
5. **Integrate with existing VHF and MPOX workflows**

This implementation provides a solid foundation for outbreak-specific access control while maintaining security and usability. 