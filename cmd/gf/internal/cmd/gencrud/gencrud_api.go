// Copyright GoFrame gf Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package gencrud

import (
	"strings"

	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"

	"github.com/gogf/gf/cmd/gf/v2/internal/utility/mlog"
)

type apiGenerator struct{}

func newApiGenerator() *apiGenerator {
	return &apiGenerator{}
}

func (g *apiGenerator) generate(table *tableInfo, dir, packageName string, overwrite bool) error {
	return g.Generate(table, dir, packageName, overwrite)
}

func (g *apiGenerator) Generate(table *tableInfo, dir, packageName string, overwrite bool) error {
	// Generate separate files for each API operation
	actions := []string{"create", "delete", "update", "get_one", "get_list", "interface"}
	
	// Create table subdirectory
	tableDir := gfile.Join(dir, gstr.CaseSnake(table.StructName))
	if !gfile.Exists(tableDir) {
		if err := gfile.Mkdir(tableDir); err != nil {
			return err
		}
	}
	
	// Create v1 subdirectory
	v1Dir := gfile.Join(tableDir, "v1")
	if !gfile.Exists(v1Dir) {
		if err := gfile.Mkdir(v1Dir); err != nil {
			return err
		}
	}

	for _, action := range actions {
		var fileName string
		var filePath string
		
		if action == "interface" {
			// 接口文件生成在表目录下
			fileName = gstr.CaseSnake(table.StructName) + ".go"
			filePath = gfile.Join(tableDir, fileName)
		} else {
			// 其他文件生成在v1目录下
			fileName = action + ".go"
			filePath = gfile.Join(v1Dir, fileName)
		}

		// Check if file exists and overwrite is false
		if gfile.Exists(filePath) && !overwrite {
			content := gfile.GetContents(filePath)
			if strings.TrimSpace(content) != "" {
				mlog.Printf("api file already exists: %s", filePath)
				continue
			}
		}

		content := g.generateActionContent(table, action)
		
		if err := gfile.PutContents(filePath, content); err != nil {
			return err
		}

		mlog.Printf("api file generated: %s", filePath)
	}

	return nil
}

func (g *apiGenerator) generateActionContent(table *tableInfo, action string) string {
	var (
		structName          = table.StructName
		structNameCaseSnake = gstr.CaseSnake(structName)
		structNameKebab     = strings.ReplaceAll(structNameCaseSnake, "_", "-")
		packageName         = "v1"
	)

	// Generate field information
	var nonPrimaryFieldReplacements []string
	var nonPrimaryRequiredFieldReplacements []string
	primaryKeyType := "uint64"
	
	for _, field := range table.Fields {
		fieldName := gstr.CaseCamel(field.Name)
		fieldType := field.Type
		jsonName := gstr.CaseCamelLower(field.Name)  // 使用小写驼峰格式的JSON字段名
		comment := field.Comment
		if comment == "" {
			comment = field.Name
		}

		if field.IsPrimaryKey {
			primaryKeyType = fieldType
		} else {
			// For update (optional fields)
			nonPrimaryFieldLine := "\t\t" + fieldName + " " + fieldType + " `json:\"" + jsonName + "\" dc:\"" + comment + "\"`"
			nonPrimaryFieldReplacements = append(nonPrimaryFieldReplacements, nonPrimaryFieldLine)
			
			// For create (required fields)
			nonPrimaryRequiredFieldLine := "\t\t" + fieldName + " " + fieldType + " `json:\"" + jsonName + "\" v:\"required\" dc:\"" + comment + "\"`"
			nonPrimaryRequiredFieldReplacements = append(nonPrimaryRequiredFieldReplacements, nonPrimaryRequiredFieldLine)
		}
	}

	var content string

	switch action {
	case "create":
		content = `// Copyright GoFrame gf Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package {{.PackageName}}

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type (
	CreateReq struct {
		g.Meta ` + "`" + `path:"/{{.StructNameKebab}}" tags:"{{.StructName}}" method:"post" summary:"Create {{.StructName}}"` + "`" + `
{{.NonPrimaryRequiredFields}}
	}
	CreateRes struct {
		g.Meta ` + "`" + `mime:"application/json" example:"string"` + "`" + `
		Id     {{.PrimaryKeyType}} ` + "`" + `json:"id" dc:"Created {{.StructNameCaseSnake}} ID"` + "`" + `
	}
)
`

	case "delete":
		content = `// Copyright GoFrame gf Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package {{.PackageName}}

import (
	"github.com/gogf/gf/v2/frame/g"
)

type (
	DeleteReq struct {
		g.Meta ` + "`" + `path:"/{{.StructNameKebab}}/{id}" tags:"{{.StructName}}" method:"delete" summary:"Delete {{.StructName}} by ID"` + "`" + `
		Id     {{.PrimaryKeyType}} ` + "`" + `json:"id" v:"required" dc:"{{.StructName}} ID"` + "`" + `
	}
	DeleteRes struct {
		g.Meta ` + "`" + `mime:"application/json" example:"string"` + "`" + `
	}
)
`

	case "update":
		content = `// Copyright GoFrame gf Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package {{.PackageName}}

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type (
	UpdateReq struct {
		g.Meta ` + "`" + `path:"/{{.StructNameKebab}}/{id}" tags:"{{.StructName}}" method:"put" summary:"Update {{.StructName}}"` + "`" + `
		Id     {{.PrimaryKeyType}} ` + "`" + `json:"id" v:"required" dc:"{{.StructName}} ID"` + "`" + `
{{.NonPrimaryFields}}
	}
	UpdateRes struct {
		g.Meta ` + "`" + `mime:"application/json" example:"string"` + "`" + `
	}
)
`

	case "get_one":
		content = `// Copyright GoFrame gf Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package {{.PackageName}}

import (
	"github.com/gogf/gf/v2/frame/g"
	"{{.ModuleName}}/internal/model/entity"
)

type (
	GetOneReq struct {
		g.Meta ` + "`" + `path:"/{{.StructNameKebab}}/{id}" tags:"{{.StructName}}" method:"get" summary:"Get {{.StructName}} by ID"` + "`" + `
		Id     {{.PrimaryKeyType}} ` + "`" + `json:"id" v:"required" dc:"{{.StructName}} ID"` + "`" + `
	}
	GetOneRes struct {
		g.Meta ` + "`" + `mime:"application/json" example:"string"` + "`" + `
		*entity.{{.StructName}} ` + "`" + `json:"{{.StructNameCaseSnake}}" dc:"{{.StructName}} info"` + "`" + `
	}
)
`

	case "get_list":
		content = `// Copyright GoFrame gf Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package {{.PackageName}}

import (
	"github.com/gogf/gf/v2/frame/g"
	"{{.ModuleName}}/internal/model/entity"
)

type (
	GetListReq struct {
		g.Meta ` + "`" + `path:"/{{.StructNameKebab}}/list" tags:"{{.StructName}}" method:"get" summary:"Get {{.StructName}} list"` + "`" + `
		Page   int ` + "`" + `d:"1" dc:"Page number"` + "`" + `
		PageSize   int ` + "`" + `d:"10" v:"max:100" dc:"Page size"` + "`" + `
		// TODO: Add other filter fields here
	}
	GetListRes struct {
		g.Meta ` + "`" + `mime:"application/json" example:"string"` + "`" + `
		List   []*entity.{{.StructName}} ` + "`" + `json:"list" dc:"{{.StructName}} list"` + "`" + `
		Total  int                      ` + "`" + `json:"total" dc:"Total count"` + "`" + `
		Page   int                      ` + "`" + `json:"page" dc:"Current page"` + "`" + `
		PageSize   int                      ` + "`" + `json:"pageSize" dc:"Page size"` + "`" + `
	}
)
`

	case "interface":
		content = `// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package {{.StructNameCaseSnake}}

import (
	"context"

	"{{.ModuleName}}/api/{{.StructNameCaseSnake}}/v1"
)

type I{{.StructName}}V1 interface {
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
	GetOne(ctx context.Context, req *v1.GetOneReq) (res *v1.GetOneRes, err error)
	GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error)
}
`
	}

	// Replace template variables
	content = gstr.ReplaceByMap(content, map[string]string{
		"{{.PackageName}}":               packageName,
		"{{.StructName}}":                structName,
		"{{.StructNameCaseSnake}}":       structNameCaseSnake,
		"{{.StructNameKebab}}":           structNameKebab,
		"{{.PrimaryKeyType}}":            primaryKeyType,
		"{{.ModuleName}}":                getModuleName(),
		"{{.NonPrimaryFields}}":          strings.Join(nonPrimaryFieldReplacements, "\n"),
		"{{.NonPrimaryRequiredFields}}":  strings.Join(nonPrimaryRequiredFieldReplacements, "\n"),
	})

	return content
}