/*Package tabs contains helper functions for managing tabs in Altid services

	go get github.com/altid/server/tabs

example

	t, _ := tabs.FromFile("/path/to/a/service")

	t.Remove("tab1")
	tab := t.Add("tab2")
	tab.SetState(Active|Alert)
	fmt.Println("%s\n", tab)

*/
package tabs
