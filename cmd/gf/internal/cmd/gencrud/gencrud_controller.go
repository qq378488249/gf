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

type controllerGenerator struct{}

func newControllerGenerator() *controllerGenerator {
	return &controllerGenerator{}
}

func NewControllerGenerator() *controllerGenerator {
	return &controllerGenerator{}
}

func (g *controllerGenerator) generate(table *tableInfo, dir, packageName string, overwrite bool) error {
	return g.Generate(table, dir, packageName, overwrite)
}

func (g *controllerGenerator) Generate(table *tableInfo, dir, packageName string, overwrite bool) error {
	// Determine package name
	if packageName == "" {
		packageName = table.Name
	}

	// Generate separate files for each CRUD operation
	actions := []string{"create", "delete", "update", "get_one", "get_list"}
	tableName := gstr.CaseSnake(table.StructName)

	// First generate the constructor file
	newFileName := tableName + "/" + tableName + "_new.go"
	newFilePath := gfile.Join(dir, newFileName)

	// Check if new file exists and overwrite is false
	if gfile.Exists(newFilePath) && !overwrite {
		content := gfile.GetContents(newFilePath)
		if strings.TrimSpace(content) != "" {
			mlog.Printf("controller new file already exists: %s", newFilePath)
		} else {
			// File exists but is empty, generate content
			newContent := g.generateNewContent(table, packageName)
			if err := gfile.PutContents(newFilePath, newContent); err != nil {
				return err
			}
			mlog.Printf("controller new file generated: %s", newFilePath)
		}
	} else {
		// File doesn't exist or overwrite is true
		newContent := g.generateNewContent(table, packageName)
		if err := gfile.PutContents(newFilePath, newContent); err != nil {
			return err
		}
		mlog.Printf("controller new file generated: %s", newFilePath)
	}

	// Then generate action files
	for _, action := range actions {
		fileName := tableName + "/" + tableName + "_v1_" + action + ".go"
		filePath := gfile.Join(dir, fileName)

		// Check if file exists and overwrite is false
		if gfile.Exists(filePath) && !overwrite {
			content := gfile.GetContents(filePath)
			if strings.TrimSpace(content) != "" {
				mlog.Printf("controller file already exists: %s", filePath)
				continue
			}
		}

		content := g.generateActionContent(table, packageName, action)

		if err := gfile.PutContents(filePath, content); err != nil {
			return err
		}

		mlog.Printf("controller file generated: %s", filePath)
	}

	return nil
}

func (g *controllerGenerator) generateActionContent(table *tableInfo, packageName, action string) string {
	var (
		structName      = table.StructName
		structNameLower = strings.ToLower(structName)
		varName         = gstr.CaseCamel(structName)
	)

	// Common header for all files
	header := `// Copyright GoFrame gf Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package {{.PackageName}}

import (
	"context"

	"{{.ModuleName}}/api/{{.StructNameLower}}/v1"
	"{{.ModuleName}}/internal/service"
)

`

	var actionContent string

	switch action {
	case "create":
		actionContent = `// Create creates a new {{.StructNameLower}} record
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	out, err := service.{{.StructName}}().Create(ctx, req)
	if err != nil {
		return nil, err
	}
	
	return out, nil
}
`

	case "delete":
		actionContent = `// Delete deletes a {{.StructNameLower}} record by ID
func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	out, err := service.{{.StructName}}().Delete(ctx, req)
	if err != nil {
		return nil, err
	}

	return out, nil
}
`

	case "update":
		actionContent = `// Update updates a {{.StructNameLower}} record
func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	out, err := service.{{.StructName}}().Update(ctx, req)
	if err != nil {
		return nil, err
	}

	return out, nil
}
`

	case "get_one":
		actionContent = `// GetOne retrieves a {{.StructNameLower}} record by ID
func (c *ControllerV1) GetOne(ctx context.Context, req *v1.GetOneReq) (res *v1.GetOneRes, err error) {
	out, err := service.{{.StructName}}().GetOne(ctx, req)
	if err != nil {
		return nil, err
	}

	return out, nil
}
`

	case "get_list":
		actionContent = `// GetList retrieves a list of {{.StructNameLower}} records
func (c *ControllerV1) GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error) {
	out, err := service.{{.StructName}}().GetList(ctx, req)
	if err != nil {
		return nil, err
	}

	return out, nil
}
`
	}

	// Combine header and action content
	content := header + actionContent

	// Replace template variables
	content = gstr.ReplaceByMap(content, map[string]string{
		"{{.PackageName}}":     packageName,
		"{{.StructName}}":      structName,
		"{{.StructNameLower}}": structNameLower,
		"{{.VarName}}":         varName,
		"{{.ModuleName}}":      getModuleName(),
	})

	return content
}

func (g *controllerGenerator) generateNewContent(table *tableInfo, packageName string) string {
	var (
		structName      = table.StructName
		structNameLower = strings.ToLower(structName)
	)

	template := `// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT. 
// =================================================================================

package {{.PackageName}}

import (
	"{{.ModuleName}}/api/{{.StructNameLower}}"
)

type ControllerV1 struct{}

func NewV1() {{.StructNameLower}}.I{{.StructName}}V1 {
	return &ControllerV1{}
}
`

	// Replace template variables
	content := template
	content = gstr.ReplaceByMap(content, map[string]string{
		"{{.PackageName}}":     packageName,
		"{{.StructName}}":      structName,
		"{{.StructNameLower}}": structNameLower,
		"{{.ModuleName}}":      getModuleName(),
	})

	return content
}
