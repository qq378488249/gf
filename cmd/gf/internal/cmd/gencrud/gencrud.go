// Copyright GoFrame gf Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package gencrud

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gtag"

	"github.com/gogf/gf/cmd/gf/v2/internal/utility/mlog"
)

const (
	CGenCrudConfig = `gfcli.gen.dao`
	CGenCrudUsage  = `gf gen crud [OPTION]`
	CGenCrudBrief  = `generate CRUD controller/service/logic files for database tables`
	CGenCrudEg     = `
gf gen crud -t user
gf gen crud -t user,product,order
gf gen crud -t user -l "mysql:root:123456@tcp(127.0.0.1:3306)/mydb"
gf gen crud -t user -c controllers -s services -o business -a api
gf gen crud -t sys_user,sys_role -r sys_
gf gen crud -t user -w
`
	CGenCrudBriefTables       = `table names to generate CRUD files for, separated by comma`
	CGenCrudBriefLink         = `database configuration, the same as the ORM configuration of GoFrame`
	CGenCrudBriefGroup        = `specifying the configuration group name of database for generated ORM instance, it's not necessary and the default value is "default"`
	CGenCrudBriefPath         = `directory path for generating files. default: ./`
	CGenCrudBriefCtrlPath     = `directory path for generating controller files. default: internal/controller`
	CGenCrudBriefServicePath  = `directory path for generating service files. default: internal/service`
	CGenCrudBriefLogicPath    = `directory path for generating logic files. default: internal/logic`
	CGenCrudBriefApiPath      = `directory path for generating api files. default: api`
	CGenCrudBriefPackageName  = `package name for generated go files`
	CGenCrudBriefOverwrite    = `overwrite all generated go files, not only empty go files`
	CGenCrudBriefRemovePrefix = `remove specified prefix of the table, multiple prefix separated with ','`
)

func init() {
	gtag.Sets(g.MapStrStr{
		`CGenCrudConfig`:            CGenCrudConfig,
		`CGenCrudUsage`:             CGenCrudUsage,
		`CGenCrudBrief`:             CGenCrudBrief,
		`CGenCrudEg`:                CGenCrudEg,
		`CGenCrudBriefTables`:       CGenCrudBriefTables,
		`CGenCrudBriefLink`:         CGenCrudBriefLink,
		`CGenCrudBriefGroup`:        CGenCrudBriefGroup,
		`CGenCrudBriefPath`:         CGenCrudBriefPath,
		`CGenCrudBriefCtrlPath`:     CGenCrudBriefCtrlPath,
		`CGenCrudBriefServicePath`:  CGenCrudBriefServicePath,
		`CGenCrudBriefLogicPath`:    CGenCrudBriefLogicPath,
		`CGenCrudBriefApiPath`:      CGenCrudBriefApiPath,
		`CGenCrudBriefPackageName`:  CGenCrudBriefPackageName,
		`CGenCrudBriefOverwrite`:    CGenCrudBriefOverwrite,
		`CGenCrudBriefRemovePrefix`: CGenCrudBriefRemovePrefix,
	})
}

type (
	CGenCrud      struct{}
	CGenCrudInput struct {
		g.Meta       `name:"crud" config:"{CGenCrudConfig}" usage:"{CGenCrudUsage}" brief:"{CGenCrudBrief}" eg:"{CGenCrudEg}"`
		Tables       string `short:"t" name:"tables" brief:"{CGenCrudBriefTables}" v:"required#please specify the table name"`
		Link         string `short:"l" name:"link" brief:"{CGenCrudBriefLink}"`
		Group        string `short:"g" name:"group" brief:"{CGenCrudBriefGroup}" d:"default"`
		Path         string `short:"p" name:"path" brief:"{CGenCrudBriefPath}" d:"./"`
		CtrlPath     string `short:"c" name:"ctrlPath" brief:"{CGenCrudBriefCtrlPath}" d:"internal/controller"`
		ServicePath  string `short:"s" name:"servicePath" brief:"{CGenCrudBriefServicePath}" d:"internal/service"`
		LogicPath    string `short:"o" name:"logicPath" brief:"{CGenCrudBriefLogicPath}" d:"internal/logic"`
		ApiPath      string `short:"a" name:"apiPath" brief:"{CGenCrudBriefApiPath}" d:"api"`
		PackageName  string `short:"n" name:"packageName" brief:"{CGenCrudBriefPackageName}"`
		Overwrite    bool   `short:"w" name:"overwrite" brief:"{CGenCrudBriefOverwrite}" orphan:"true"`
		RemovePrefix string `short:"r" name:"removePrefix" brief:"{CGenCrudBriefRemovePrefix}"`
	}
	CGenCrudOutput struct{}
)

type TableInfo struct {
	Name       string
	StructName string
	Fields     []FieldInfo
	PrimaryKey string
}

type FieldInfo struct {
	Name         string
	Type         string
	JsonName     string
	Comment      string
	IsPrimaryKey bool
}

// For backward compatibility
type tableInfo = TableInfo
type fieldInfo = FieldInfo

func (c CGenCrud) Crud(ctx context.Context, in CGenCrudInput) (out *CGenCrudOutput, err error) {
	// Support config file like dao command
	if in.Link != "" {
		c.doGenCrudForLink(ctx, in)
	} else if g.Cfg().Available(ctx) {
		v := g.Cfg().MustGet(ctx, CGenCrudConfig)
		if v.IsSlice() {
			for i := 0; i < len(v.Interfaces()); i++ {
				c.doGenCrudForArray(ctx, i, in)
			}
		} else {
			c.doGenCrudForArray(ctx, -1, in)
		}
	} else {
		c.doGenCrudForLink(ctx, in)
	}

	mlog.Print("done!")
	return
}

// doGenCrudForArray implements the "gen crud" command for configuration array.
func (c CGenCrud) doGenCrudForArray(ctx context.Context, index int, in CGenCrudInput) {
	var err error
	if index >= 0 {
		err = g.Cfg().MustGet(
			ctx,
			fmt.Sprintf(`%s.%d`, CGenCrudConfig, index),
		).Scan(&in)
		if err != nil {
			mlog.Fatalf(`invalid configuration of "%s": %+v`, CGenCrudConfig, err)
		}
	}
	c.doGenCrudForLink(ctx, in)
}

// doGenCrudForLink implements the actual CRUD generation logic.
func (c CGenCrud) doGenCrudForLink(ctx context.Context, in CGenCrudInput) {
	// Get database connection using dao command logic
	db, err := c.getDatabase(in.Link, in.Group)
	if err != nil {
		mlog.Fatalf(`database initialization failed: %+v`, err)
	}

	// Parse table names
	tableNames := gstr.SplitAndTrim(in.Tables, ",")
	if len(tableNames) == 0 {
		mlog.Fatal("table names cannot be empty")
	}

	// Process each table
	for _, tableName := range tableNames {
		if tableName == "" {
			continue
		}

		// Get table info
		table, err := c.getTableInfo(db, tableName, in.RemovePrefix)
		if err != nil {
			mlog.Fatalf(`get table info failed for table "%s": %+v`, tableName, err)
		}

		// Generate files
		if err := c.generateFiles(table, in); err != nil {
			mlog.Fatalf(`generate files failed for table "%s": %+v`, tableName, err)
		}

		mlog.Printf("CRUD files generated successfully for table: %s", tableName)
	}
}

func (c CGenCrud) getDatabase(link, group string) (gdb.DB, error) {
	var (
		db  gdb.DB
		err error
	)

	// It uses user passed database configuration.
	if link != "" {
		var tempGroup = gtime.TimestampNanoStr()
		err = gdb.AddConfigNode(tempGroup, gdb.ConfigNode{
			Link: link,
		})
		if err != nil {
			return nil, gerror.Newf(`database configuration failed: %+v`, err)
		}
		if db, err = gdb.Instance(tempGroup); err != nil {
			return nil, gerror.Newf(`database initialization failed: %+v`, err)
		}
	} else {
		db = g.DB(group)
	}
	if db == nil {
		if link == "" {
			return nil, gerror.New(`database initialization failed: please provide database configuration via -l parameter or config file. Example: gf gen crud -t user -l "mysql:root:123456@tcp(127.0.0.1:3306)/dbname"`)
		}
		return nil, gerror.New(`database initialization failed, may be invalid database configuration`)
	}

	return db, nil
}

func (c CGenCrud) getTableInfo(db gdb.DB, tableName, removePrefix string) (*tableInfo, error) {
	// Remove prefix if specified
	structName := tableName
	if removePrefix != "" {
		prefixes := gstr.SplitAndTrim(removePrefix, ",")
		for _, prefix := range prefixes {
			if gstr.HasPrefix(structName, prefix) {
				structName = gstr.TrimLeft(structName, prefix)
				break
			}
		}
	}

	// Convert table name to struct name (CamelCase)
	structName = gstr.CaseCamel(structName)

	// Get table fields
	fields, err := db.TableFields(context.TODO(), tableName)
	if err != nil {
		return nil, err
	}

	table := &tableInfo{
		Name:       tableName,
		StructName: structName,
		Fields:     make([]fieldInfo, 0),
	}

	// Process fields
	for fieldName, field := range fields {
		fieldInfo := fieldInfo{
			Name:     fieldName,
			JsonName: gstr.CaseSnake(fieldName),
			Comment:  field.Comment,
		}

		// Determine Go type based on database type
		fieldInfo.Type = c.getGoTypeFromDbType(field.Type)

		// Check if primary key
		if gstr.Contains(field.Key, "PRI") {
			fieldInfo.IsPrimaryKey = true
			table.PrimaryKey = fieldName
		}

		table.Fields = append(table.Fields, fieldInfo)
	}

	if table.PrimaryKey == "" {
		// Use 'id' as default primary key if no primary key found
		table.PrimaryKey = "id"
	}

	return table, nil
}

func (c CGenCrud) getGoTypeFromDbType(dbType string) string {
	dbType = strings.ToLower(dbType)

	switch {
	case gstr.Contains(dbType, "int"):
		if gstr.Contains(dbType, "tinyint") {
			return "int8"
		} else if gstr.Contains(dbType, "smallint") {
			return "int16"
		} else if gstr.Contains(dbType, "mediumint") {
			return "int32"
		} else if gstr.Contains(dbType, "bigint") {
			return "int64"
		}
		return "int"
	case gstr.Contains(dbType, "decimal") || gstr.Contains(dbType, "numeric"):
		return "float64"
	case gstr.Contains(dbType, "float"):
		return "float32"
	case gstr.Contains(dbType, "double"):
		return "float64"
	case gstr.Contains(dbType, "char") || gstr.Contains(dbType, "text") || gstr.Contains(dbType, "blob"):
		return "string"
	case gstr.Contains(dbType, "date") || gstr.Contains(dbType, "time"):
		return "*gtime.Time"
	case gstr.Contains(dbType, "json"):
		return "string"
	default:
		return "string"
	}
}

func (c CGenCrud) generateFiles(table *tableInfo, in CGenCrudInput) error {
	// Ensure directories exist
	ctrlDir := gfile.Join(in.Path, in.CtrlPath)
	serviceDir := gfile.Join(in.Path, in.ServicePath)
	logicDir := gfile.Join(in.Path, in.LogicPath, gstr.CaseSnake(table.StructName))
	apiDir := gfile.Join(in.Path, in.ApiPath)

	if !gfile.Exists(ctrlDir) {
		if err := gfile.Mkdir(ctrlDir); err != nil {
			return err
		}
	}
	if !gfile.Exists(serviceDir) {
		if err := gfile.Mkdir(serviceDir); err != nil {
			return err
		}
	}
	if !gfile.Exists(logicDir) {
		if err := gfile.Mkdir(logicDir); err != nil {
			return err
		}
	}
	if !gfile.Exists(apiDir) {
		if err := gfile.Mkdir(apiDir); err != nil {
			return err
		}
	}

	// Generate API files
	if err := newApiGenerator().generate(table, apiDir, "v1", in.Overwrite); err != nil {
		return err
	}

	// Generate controller files
	if err := newControllerGenerator().generate(table, ctrlDir, in.PackageName, in.Overwrite); err != nil {
		return err
	}

	// Generate service file
	if err := newServiceGenerator().generate(table, serviceDir, in.PackageName, in.Overwrite); err != nil {
		return err
	}

	// Generate logic file
	if err := newLogicGenerator().generate(table, logicDir, in.PackageName, in.Overwrite); err != nil {
		return err
	}

	return nil
}

// getModuleName detects module name from go.mod file
func getModuleName() string {
	// Try to detect module name from go.mod
	if gfile.Exists("go.mod") {
		content := gfile.GetContents("go.mod")
		lines := gstr.SplitAndTrim(content, "\n")
		if len(lines) > 0 {
			firstLine := lines[0]
			if gstr.HasPrefix(firstLine, "module ") {
				return gstr.Trim(firstLine[7:])
			}
		}
	}
	return "your-module-name"
}
