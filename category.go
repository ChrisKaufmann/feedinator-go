package main

type Category struct {
	Name        string
	Description string
	UserName    string
	ID          int
	Unread      int
	Evenodd     string
	Class       string
}

func getCat(id string) Category {
	var cat Category
	err := stmtGetCat.QueryRow(id).Scan(&cat.Name, &cat.UserName, &cat.Description, &cat.ID)
	if err != nil {
		err.Error()
	}
	return cat
}
func getCategoryFeeds(id string) []Feed {
	var allFeeds []Feed
	rows, err := stmtGetCatFeeds.Query(id)
	if err != nil {
		err.Error()
		return allFeeds
	}
	for rows.Next() {
		var id string
		rows.Scan(&id)
		allFeeds = append(allFeeds, getFeed(id))
	}
	return allFeeds
}
func getCategories() []Category {
	var allCats []Category
	rows, err := stmtGetCats.Query(userName)
	if err != nil {
		err.Error()
		return allCats
	}
	for rows.Next() {
		var cat Category
		rows.Scan(&cat.Name, &cat.UserName, &cat.Description, &cat.ID)
		cat.Unread = unreadCategoryCount(cat.ID)
		if cat.Unread > 0 {
			cat.Class = "oddUnread"
		} else {
			cat.Class = "odd"
		}
		allCats = append(allCats, cat)
	}
	return allCats
}
func unreadCategoryCount(id int) int {
	var count int
	err := stmtCatUnread.QueryRow(id).Scan(&count)
	if err != nil {
		err.Error()
	}
	return count
}
