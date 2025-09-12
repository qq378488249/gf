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

type serviceGenerator struct{}

func newServiceGenerator() *serviceGenerator {
	return &serviceGenerator{}
}

func NewServiceGenerator() *serviceGenerator {
	return &serviceGenerator{}
}

func (g *serviceGenerator) generate(table *tableInfo, dir, packageName string, overwrite bool) error {
	return g.Generate(table, dir, packageName, overwrite)
}

func (g *serviceGenerator) Generate(table *tableInfo, dir, packageName string, overwrite bool) error {
	fileName := gstr.CaseSnake(table.StructName) + ".go"
	filePath := gfile.Join(dir, fileName)

	// Check if file exists and overwrite is false
	if gfile.Exists(filePath) && !overwrite {
		content := gfile.GetContents(filePath)
		if strings.TrimSpace(content) != "" {
			mlog.Printf("service file already exists: %s", filePath)
			return nil
		}
	}

	// Determine package name
	if packageName == "" {
		packageName = gfile.Basename(dir)
	}

	content := g.generateContent(table, packageName)
	
	if err := gfile.PutContents(filePath, content); err != nil {
		return err
	}

	mlog.Printf("service file generated: %s", filePath)
	return nil
}

func (g *serviceGenerator) generateContent(table *tableInfo, packageName string) string {
	var (
		structName      = table.StructName
		structNameLower = strings.ToLower(structName)
	)

	template := `// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package {{.PackageName}}

import (
	"context"
	v1 "{{.ModuleName}}/api/{{.StructNameLower}}/v1"
)

type (
	I{{.StructName}} interface {
		// Create creates a new {{.StructNameLower}} record
		Create(ctx context.Context, req *v1.CreateReq) (*v1.CreateRes, error)
		// Delete deletes {{.StructNameLower}} record by ID
		Delete(ctx context.Context, req *v1.DeleteReq) (*v1.DeleteRes, error)
		// Update updates {{.StructNameLower}} record
		Update(ctx context.Context, req *v1.UpdateReq) (*v1.UpdateRes, error)
		// GetOne retrieves {{.StructNameLower}} record by ID
		GetOne(ctx context.Context, req *v1.GetOneReq) (*v1.GetOneRes, error)
		// GetList retrieves {{.StructNameLower}} records with pagination
		GetList(ctx context.Context, req *v1.GetListReq) (*v1.GetListRes, error)
	}
)

var (
	local{{.StructName}} I{{.StructName}}
)

func {{.StructName}}() I{{.StructName}} {
	if local{{.StructName}} == nil {
		panic("implement not found for interface I{{.StructName}}, forgot register?")
	}
	return local{{.StructName}}
}

func Register{{.StructName}}(i I{{.StructName}}) {
	local{{.StructName}} = i
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