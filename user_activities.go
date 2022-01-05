package passport

import (
	"encoding/json"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
)

// UserActivity is a single userActivity on the platform
type UserActivity struct {
	ID         UserActivityID `json:"id" db:"id"`
	UserID     UserID         `json:"userID" db:"user_id"`
	Action     string         `json:"action" db:"action"`
	ObjectID   *string        `json:"objectID,omitempty" db:"object_id"`
	ObjectSlug *string        `json:"objectSlug,omitempty" db:"object_slug"`
	ObjectName *string        `json:"objectName,omitempty" db:"object_name"`
	ObjectType ObjectType     `json:"objectType" db:"object_type"`
	OldData    null.JSON      `json:"oldData,omitempty" db:"old_data"`
	NewData    null.JSON      `json:"newData,omitempty" db:"new_data"`
	CreatedAt  time.Time      `json:"createdAt" db:"created_at"`
	User       *User          `json:"user" db:"user"`
}

// UserActivityChangeData contains a set of data from the db to compare with
type UserActivityChangeData struct {
	Name string
	From interface{}
	To   interface{}
}

// ObjectType enum used for user activity logging
type ObjectType string

// ObjectType enums
const (
	ObjectTypeBlob         ObjectType = "Blob"
	ObjectTypeOrganisation ObjectType = "Organisation"
	ObjectTypeRole         ObjectType = "Role"
	ObjectTypeUser         ObjectType = "User"
	ObjectTypeProduct      ObjectType = "Product"
)

// AllObjectType contains all ObjectType enums
var AllObjectType = []ObjectType{
	ObjectTypeBlob,
	ObjectTypeOrganisation,
	ObjectTypeRole,
	ObjectTypeUser,
}

func (e ObjectType) String() string {
	return string(e)
}

// UserActivityGetDataChanges returns oldData and newData JSON
func UserActivityGetDataChanges(changes []*UserActivityChangeData) (null.JSON, null.JSON, error) {
	oldData := make(map[string]interface{})
	newData := make(map[string]interface{})

	oldDataJson := null.JSONFromPtr(nil)
	newDataJson := null.JSONFromPtr(nil)

	for _, c := range changes {
		if c.From != nil {
			oldData[c.Name] = c.From
		}
		if c.To != nil {
			newData[c.Name] = c.To
		}
	}

	data, err := json.Marshal(oldData)
	if err != nil {
		return oldDataJson, newDataJson, terror.Error(err, "Failed to record user activity")
	}
	oldDataJson = null.JSONFrom(data)

	data, err = json.Marshal(newData)
	if err != nil {
		return oldDataJson, newDataJson, terror.Error(err, "Failed to record user activity")
	}
	newDataJson = null.JSONFrom(data)

	return oldDataJson, newDataJson, nil
}
