package sqldata

var (
	selectGroupTemplatesSQL = "SELECT group_id, template_id FROM group_template ORDER BY group_id ASC, template_id ASC"
)

// ReadGroupTemplates query template ids for host-group
func ReadGroupTemplates() (map[int][]int, error) {
	rows, err := portalDB.Query(selectGroupTemplatesSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var (
		templates = make(map[int][]int)
		count     int
	)
	for rows.Next() {
		var groupID int
		var templateID int
		if err = rows.Scan(&groupID, &templateID); err != nil {
			continue
		}
		templates[groupID] = append(templates[groupID], templateID)
		count++
	}
	return templates, nil
}

var selectGroupsSQL = "SELECT id,group_name FROM `group`"

type groupItem struct {
	ID   int    `db:"id"`
	Name string `db:"group_name"`
}

// ReadGroupNames query groups data
func ReadGroupNames() (map[int]string, error) {
	rows, err := portalDB.Queryx(selectGroupsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var (
		groups = make(map[int]string)
		count  int
	)
	for rows.Next() {
		gp := groupItem{}
		if err = rows.StructScan(&gp); err != nil {
			continue
		}
		if gp.ID > 0 {
			groups[gp.ID] = gp.Name
			count++
		}
	}
	return groups, nil
}
