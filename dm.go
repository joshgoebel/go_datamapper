package dm

import (
	"sqlite3";
	"fmt";
	"strings";
	"reflect";
	"container/vector";
	"os";
	)

type Connection sqlite3.Handle;

type Model struct {
	table_name string;
	connection Connection;
}

type ResultSet struct {
	Results *vector.Vector;
}

type Opts map[string] interface{};

// func main()
// {
// 	
// }

const (
	SQLITE_INTEGER = 1;
	SQLITE_TEXT = 3;
	SQLITE_ROW = 100;
	)

var conn *sqlite3.Handle;
var model_map map[string] reflect.Type;

func Init(dbname string) {
	conn = new(sqlite3.Handle);
	model_map = make(map[string] reflect.Type);
	r := conn.Open(dbname);
	if r!="" {
		println("ERROR");
	} 
	
}

func RegisterModel(model string, t reflect.Type) {
	model_map[model]=t;
}

func Find(model string, id int) interface{}
{
	sql := select_one_sql(model, id);
	return single_result(sql, model);
}

func FindLast(model string) interface{} {
	sql := select_all_sql(model) + " order by id desc limit 1";
	return single_result(sql, model);
}

func FindFirst(model string) interface{} {
	sql := select_all_sql(model) + " limit 1";
	return single_result(sql, model);
}

func Count(model string) int {
	sql := select_count_sql(model);
	res, err := Execute(sql);
	if err!=0 {
		defer res.Finalize();
	}
	return res.ColumnInt(0);
}

func single_result(sql string, model string) interface{} {
	res, err := Execute(sql);
	if err!=0 {
		defer res.Finalize();
	}	
	o := reflect.MakeZero(model_map[model]);
	build_result(o, res);	
	return o.Interface();
}

func parse_options(opts ...) (options Opts) {
	v := reflect.NewValue(opts).(*reflect.StructValue);
	if v.NumField() > 0 {
		options = v.Field(0).Interface().(Opts);
	} else {
		options = Opts{};
	}
	return;
}

func FindAll(model string, opts ...) (r ResultSet) {
	options := parse_options(opts);
	r = ResultSet{};
	sql := "";
	limit, ok := options["limit"];
	if ok && limit.(int) > 0 {
		sql = select_all_sql(model) + fmt.Sprintf (" limit %d", options["limit"].(int))
	}
	else
	{
		sql = select_all_sql(model);
	}
	res, err := Execute(sql);
	if err!=0 {
		defer res.Finalize();
	}
	r.Results = build_results(model, res);	
	return;
}

func camel_case(s string) string {
	return strings.ToUpper(s[0:1]) + s[1:len(s)]
}

func build_results(model string, st *sqlite3.Statement) (r *vector.Vector)
{
	r = vector.New(0);
	for i := SQLITE_ROW; i == SQLITE_ROW; i= st.Step() {
		o := reflect.MakeZero(model_map[model]);
		x := build_result(o, st);
		r.Push(x.Interface());
	}
	return;
}

func build_result(o reflect.Value, st *sqlite3.Statement ) reflect.Value
{
	cc := st.ColumnCount();
	for i:=0; i<cc; i++ {
		cn := camel_case(st.ColumnName(i));
		ct := st.ColumnType(i);
		f := o.(*reflect.StructValue).FieldByName(cn);
		switch ct {
		case SQLITE_INTEGER: f.(*reflect.IntValue).Set(st.ColumnInt(i));
		case SQLITE_TEXT: f.(*reflect.StringValue).Set(st.ColumnText(i));
		}	
	}
	return o;
}

func Execute(sql string) (s *sqlite3.Statement, err int)
{
	errs := "";
	// println(sql);
	s, errs = conn.Prepare(sql);
	if errs!="" {
		println("SQL ERROR: " + conn.ErrMsg());
		die();
	}
	err = s.Step();
	return s, err;
}

func die() {
	os.Exit(1);
}

func select_one_sql(model string, id int) string {
	return fmt.Sprintf("select * from " + model + "s where id = %d", id);
}

func select_all_sql(model string) string {
	return fmt.Sprintf("select * from " + model + "s");
}

func select_count_sql(model string) string {
	return fmt.Sprintf("select count(id) from " + model + "s");
}