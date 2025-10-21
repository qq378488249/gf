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

type logicGenerator struct{}

func newLogicGenerator() *logicGenerator {
	return &logicGenerator{}
}

func NewLogicGenerator() *logicGenerator {
	return &logicGenerator{}
}

func (g *logicGenerator) generate(table *tableInfo, dir, packageName string, overwrite bool) error {
	return g.Generate(table, dir, packageName, overwrite)
}

func (g *logicGenerator) Generate(table *tableInfo, dir, packageName string, overwrite bool) error {
	fileName := gstr.CaseSnake(table.StructName) + ".go"
	filePath := gfile.Join(dir, fileName)

	// Check if file exists and overwrite is false
	if gfile.Exists(filePath) && !overwrite {
		content := gfile.GetContents(filePath)
		if strings.TrimSpace(content) != "" {
			mlog.Printf("logic file already exists: %s", filePath)
			return nil
		}
	}

	// Determine package name
	if packageName == "" {
		packageName = gstr.CaseSnake(table.StructName)
	}

	content := g.generateContent(table, packageName)
	
	if err := gfile.PutContents(filePath, content); err != nil {
		return err
	}

	mlog.Printf("logic file generated: %s", filePath)
	return nil
}

func (g *logicGenerator) generateContent(table *tableInfo, packageName string) string {
	var (
		structName          = table.StructName
		structNameCaseSnake = gstr.CaseSnake(structName)
	)

	template := `// Copyright GoFrame gf Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package {{.PackageName}}

import (
	"context"

	"{{.ModuleName}}/api/{{.StructNameCaseSnake}}/v1"
	"{{.ModuleName}}/internal/dao"
	"{{.ModuleName}}/internal/model/entity"
	"{{.ModuleName}}/internal/service"
)

type (
	s{{.StructName}} struct{}
)

func init() {
	service.Register{{.StructName}}(New())
}

func New() service.I{{.StructName}} {
	return &s{{.StructName}}{}
}

// Create creates a new {{.StructNameCaseSnake}} record
func (s *s{{.StructName}}) Create(ctx context.Context, req *v1.CreateReq) (*v1.CreateRes, error) {
	// TODO: Add business logic here before creating the record
	// For example: validation, data transformation, etc.

	// Insert into database
	result, err := dao.{{.StructName}}.Ctx(ctx).Data(req).Insert()
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &v1.CreateRes{
		Id: uint64(id),
	}, nil
}

// Delete deletes a {{.StructNameCaseSnake}} record by ID
func (s *s{{.StructName}}) Delete(ctx context.Context, req *v1.DeleteReq) (*v1.DeleteRes, error) {
	// TODO: Add business logic here before deleting the record
	// For example: permission checks, related data cleanup, etc.

	_, err := dao.{{.StructName}}.Ctx(ctx).Where(dao.{{.StructName}}.Columns().Id, req.Id).Delete()
	if err != nil {
		return nil, err
	}

	return &v1.DeleteRes{}, nil
}

// Update updates a {{.StructNameCaseSnake}} record
func (s *s{{.StructName}}) Update(ctx context.Context, req *v1.UpdateReq) (*v1.UpdateRes, error) {
	// TODO: Add business logic here before updating the record
	// For example: validation, data transformation, permission checks, etc.

	_, err := dao.{{.StructName}}.Ctx(ctx).Data(req).Where(dao.{{.StructName}}.Columns().Id, req.Id).Update()
	if err != nil {
		return nil, err
	}

	return &v1.UpdateRes{}, nil
}

// GetOne retrieves a {{.StructNameCaseSnake}} record by ID
func (s *s{{.StructName}}) GetOne(ctx context.Context, req *v1.GetOneReq) (*v1.GetOneRes, error) {
	// TODO: Add business logic here before getting the record
	// For example: permission checks, data filtering, etc.

	var entity *entity.{{.StructName}}
	err := dao.{{.StructName}}.Ctx(ctx).Where(dao.{{.StructName}}.Columns().Id, req.Id).Scan(&entity)
	if err != nil {
		return nil, err
	}

	return &v1.GetOneRes{
		{{.StructName}}: entity,
	}, nil
}

// GetList retrieves a list of {{.StructNameCaseSnake}} records
func (s *s{{.StructName}}) GetList(ctx context.Context, req *v1.GetListReq) (*v1.GetListRes, error) {
	// TODO: Add business logic here before getting the list
	// For example: permission checks, data filtering, search conditions, etc.

	var (
		entities []*entity.{{.StructName}}
		model    = dao.{{.StructName}}.Ctx(ctx)
	)

	// Add conditions if provided
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// Apply pagination
	model = model.Page(req.Page, req.PageSize)

	// Execute query
	err := model.Scan(&entities)
	if err != nil {
		return nil, err
	}

	// Get total count
	total, err := dao.{{.StructName}}.Ctx(ctx).Count()
	if err != nil {
		return nil, err
	}

	return &v1.GetListRes{
		List:     entities,
		Page:     req.Page,
		PageSize: req.PageSize,
		Total:    total,
	}, nil
}
`

	// Replace template variables
	content := template
	content = gstr.ReplaceByMap(content, map[string]string{
		"{{.PackageName}}":         packageName,
		"{{.StructName}}":          structName,
		"{{.StructNameCaseSnake}}": structNameCaseSnake,
		"{{.ModuleName}}":          getModuleName(),
	})

	return content
}