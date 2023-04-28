package storage

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type JSONQueryExpression struct {
	column string
	keys   []string

	not    bool
	values []string
}

func JSONQuery(column string, keys ...string) *JSONQueryExpression {
	return &JSONQueryExpression{column: column, keys: keys}
}

func (jsonQuery *JSONQueryExpression) Exist() *JSONQueryExpression {
	jsonQuery.not = false
	return jsonQuery
}

func (jsonQuery *JSONQueryExpression) NotExist() *JSONQueryExpression {
	jsonQuery.not = true
	return jsonQuery
}

func (jsonQuery *JSONQueryExpression) Equal(value string) *JSONQueryExpression {
	jsonQuery.not, jsonQuery.values = false, []string{value}
	return jsonQuery
}

func (jsonQuery *JSONQueryExpression) NotEqual(value string) *JSONQueryExpression {
	jsonQuery.not, jsonQuery.values = true, []string{value}
	return jsonQuery
}

func (jsonQuery *JSONQueryExpression) In(values ...string) *JSONQueryExpression {
	jsonQuery.not, jsonQuery.values = false, values
	return jsonQuery
}

func (jsonQuery *JSONQueryExpression) NotIn(values ...string) *JSONQueryExpression {
	jsonQuery.not, jsonQuery.values = true, values
	return jsonQuery
}

func (jsonQuery *JSONQueryExpression) Build(builder clause.Builder) {
	if len(jsonQuery.keys) == 0 {
		return
	}

	if stmt, ok := builder.(*gorm.Statement); ok {
		switch stmt.Dialector.Name() {
		case "arango":
			query := strings.Join(jsonQuery.keys[0:len(jsonQuery.keys)-1], ".")
			query = query + fmt.Sprintf("['%s']", jsonQuery.keys[len(jsonQuery.keys)-1])
			value := jsonQuery.values[0]
			query = fmt.Sprintf(" doc.object.%s == '%s' FILTER ", query, value)
			writeString(builder, query)
		}
	}
}

func writeString(builder clause.Writer, str string) {
	_, _ = builder.WriteString(str)
}

func buildOwnerQueryByUID(db *gorm.DB, cluster, uid string, seniority int) interface{} {
	if seniority == 0 {
		return uid
	}

	parentOwner := buildOwnerQueryByUID(db, cluster, uid, seniority-1)
	ownerQuery := db.Model(Resource{}).Select("uid").Where(map[string]interface{}{"cluster": cluster})
	if _, ok := parentOwner.(string); ok {
		return ownerQuery.Where("owner_uid = ?", parentOwner)
	}
	return ownerQuery.Where("owner_uid IN (?)", parentOwner)
}

func buildOwnerQueryByName(db *gorm.DB, cluster string, namespaces []string, groupResource schema.GroupResource, name string, seniority int) interface{} {
	ownerQuery := db.Model(Resource{}).Select("uid").Where(map[string]interface{}{"cluster": cluster})
	if seniority != 0 {
		parentOwner := buildOwnerQueryByName(db, cluster, namespaces, groupResource, name, seniority-1)
		return ownerQuery.Where("owner_uid IN (?)", parentOwner)
	}

	if !groupResource.Empty() {
		ownerQuery = ownerQuery.Where(map[string]interface{}{"group": groupResource.Group, "resource": groupResource.Resource})
	}
	switch len(namespaces) {
	case 0:
	case 1:
		ownerQuery = ownerQuery.Where("namespace = ?", namespaces[0])
	default:
		ownerQuery = ownerQuery.Where("namespace IN (?)", namespaces)
	}
	return ownerQuery.Where("name = ?", name)
}
