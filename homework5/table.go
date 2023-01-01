package orm

type TableReference interface {
	tableAlias() string
}

type Table struct {
	entity any
	alias  string
	cols   []Selectable
}

func TableOf(entity any) Table {
	return Table{
		entity: entity,
	}
}

func (t Table) C(name string) Column {
	return Column{
		name:  name,
		table: t,
	}
}

func (t Table) tableAlias() string {
	return t.alias
}

func (t Table) Checkin(name string) bool {
	//fmt.Println("checkin", name)
	//fmt.Println("checkin", name, t.cols)
	//做校验，拿到的name 是否包含在selectable
	if len(t.cols) > 0 {
		//fmt.Println("我需要有单独指定的列明", name)
		for _, col := range t.cols {
			if col.fieldName() == name {
				return true
			}
		}
		return false
	}
	return true
}

func (t Table) As(alias string) Table {
	return Table{
		entity: t.entity,
		alias:  alias,
	}
}

func (t Table) Add(name ...Selectable) Table {
	return Table{
		entity: t.entity,
		alias:  t.alias,
		cols:   append(t.cols, name...),
	}
}
func (t Table) Join(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: target,
		typ:   "JOIN",
	}
}

func (t Table) LeftJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: target,
		typ:   "LEFT JOIN",
	}
}

func (t Table) RightJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: target,
		typ:   "RIGHT JOIN",
	}
}

type JoinBuilder struct {
	left  TableReference
	right TableReference
	typ   string
}

var _ TableReference = Join{}

type Join struct {
	left  TableReference
	right TableReference
	typ   string
	on    []Predicate
	using []string
}

func (j Join) Join(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: target,
		typ:   "JOIN",
	}
}

func (j Join) LeftJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: target,
		typ:   "LEFT JOIN",
	}
}

func (j Join) RightJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: target,
		typ:   "RIGHT JOIN",
	}
}

func (j Join) tableAlias() string {
	return ""
}

func (j *JoinBuilder) On(ps ...Predicate) Join {
	return Join{
		left:  j.left,
		right: j.right,
		on:    ps,
		typ:   j.typ,
	}
}

func (j *JoinBuilder) Using(cs ...string) Join {
	return Join{
		left:  j.left,
		right: j.right,
		using: cs,
		typ:   j.typ,
	}
}

type Subquery struct {
	s       *Selector[any]
	entity  any
	alias   string
	error   error
	columns []Selectable
}

func (s Subquery) expr() {}

func (s Subquery) tableAlias() string {
	return s.alias
}

func (s Subquery) Join(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  s,
		right: target,
		typ:   "JOIN",
	}
}

func (s Subquery) LeftJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  s,
		right: target,
		typ:   "LEFT JOIN",
	}
}

func (s Subquery) RightJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  s,
		right: target,
		typ:   "RIGHT JOIN",
	}
}

func (s Subquery) C(name string) Column {
	//fmt.Println(s.s.model)
	//for _, values := range s.columns {
	//	//fmt.Println(index, values)
	//	//fmt.Println("TAGMMM", name, values.target(), values.fieldName(), values.selectedAlias())
	//}

	//有指明的话
	res := Column{table: TableOf(s.entity).As(s.alias).Add(s.columns...), name: name}
	//fmt.Println(res)
	//res.Makealias()

	return res
}
