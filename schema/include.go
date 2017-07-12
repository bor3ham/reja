package schema

type Include struct {
	Children map[string]*Include
}

func (i *Include) endNodes(prefix string) []string {
	var nodes []string
	for key, child := range i.Children {
		if len(child.Children) > 0 {
			nodes = append(nodes, child.endNodes(prefix + key + ".")...)
		} else {
			nodes = append(nodes, prefix + key)
		}
	}
	return nodes
}

func (i *Include) AsString() string {
	query := ""
	endNodes := i.endNodes("")
	for index, node := range endNodes {
		if index != 0 {
			query += ","
		}
		query += node
	}
	return query
}
