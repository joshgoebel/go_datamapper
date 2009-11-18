package dm

import (
	"sqlite3";
	"fmt";
	"strings";
	"reflect";
	"container/vector";
	"os";
)

//type Connection sqlite3.Handle;

type Model struct {
	Name		string;
	TableName	string;
	Type		reflect.Type;
	Connection	* sqlite3.Handle;
}

type ResultSet struct {
	Results *vector.Vector;
}

type Opts map[string]interface{}

// func main()
// {
//
// }

const (
	SQLITE_INTEGER	= 1;
	SQLITE_TEXT	= 3;
	SQLITE_ROW	= 100;
)

var conn *sqlite3.Handle
var models map[string]Model

func AddModel(name string, table string, t reflect.Type) (m Model){
	m = Model{name, table, t, conn};
	models[name] = m;
	return
}

func Init(dbname string) {
	conn = new(sqlite3.Handle);
	models = make(map[string] Model);
	r := conn.Open(dbname);
	if r != "" {
		println("ERROR")
	}

}

func (m *Model) single_result(sql string) interface{} {
	res, err := Execute(sql);
	if err != 0 {
		defer res.Finalize()
	}
	o := reflect.MakeZero(m.Type);
	build_result(o, res);
	return o.Interface();
}

func parse_options(opts ...) (options Opts) {
	v := reflect.NewValue(opts).(*reflect.StructValue);
	if v.NumField() > 0 {
		options = v.Field(0).Interface().(Opts)
	} else {
		options = Opts{}
	}
	return;
}

func camel_case(s string) string	{ return strings.ToUpper(s[0:1]) + s[1:len(s)] }

func (m *Model) build_results(st *sqlite3.Statement) (r *vector.Vector) {
	r = vector.New(0);
	for i := SQLITE_ROW; i == SQLITE_ROW; i = st.Step() {
		o := reflect.MakeZero(m.Type);
		x := build_result(o, st);
		r.Push(x.Interface());
	}
	return;
}

func build_result(o reflect.Value, st *sqlite3.Statement) reflect.Value {
	cc := st.ColumnCount();
	for i := 0; i < cc; i++ {
		cn := camel_case(st.ColumnName(i));
		ct := st.ColumnType(i);
		f := o.(*reflect.StructValue).FieldByName(cn);
		switch ct {
		case SQLITE_INTEGER:
			f.(*reflect.IntValue).Set(st.ColumnInt(i))
		case SQLITE_TEXT:
			f.(*reflect.StringValue).Set(st.ColumnText(i))
		}
	}
	return o;
}

func Execute(sql string) (s *sqlite3.Statement, err int) {
	errs := "";
	// println(sql);
	s, errs = conn.Prepare(sql);
	if errs != "" {
		println("SQL ERROR: " + conn.ErrMsg());
		die();
	}
	err = s.Step();
	return s, err;
}

func die()	{ os.Exit(1) }

func select_one_sql(model string, id int) string {
	return fmt.Sprintf("select * from "+model+"s where id = %d", id)
}

func select_all_sql(model string) string	{ 
	return fmt.Sprintf("select * from " + model + "s") 
}

func select_count_sql(model string) string {
	return fmt.Sprintf("select count(id) from " + model + "s")
}

// model stuff

func (m *Model) Count() int {
	sql := select_count_sql(m.Name);
	res, err := Execute(sql);
	if err != 0 {
		defer res.Finalize()
	}
	return res.ColumnInt(0);
}

func (m *Model) All(opts ...) (r ResultSet) {
	options := parse_options(opts);
	r = ResultSet{};
	sql := "";
	limit, ok := options["limit"];
	if ok && limit.(int) > 0 {
		sql = select_all_sql(m.Name) + fmt.Sprintf(" limit %d", options["limit"].(int))
	} else {
		sql = select_all_sql(m.Name)
	}
	res, err := Execute(sql);
	if err != 0 {
		defer res.Finalize()
	}
	r.Results = m.build_results(res);
	return;
}

func (m *Model) First() interface{} {
	sql := select_all_sql(m.Name) + " limit 1";
	return m.single_result(sql);	
}

func (m *Model) Find(id int) interface{} {
	sql := select_one_sql(m.Name, id);
	return m.single_result(sql);
}

func (m *Model) Last() interface{} {
	sql := select_all_sql(m.Name) + " order by id desc limit 1";
	return m.single_result(sql);
}
