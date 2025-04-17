package main

func main() {
	username, address := ShowConnectionDialog()
	client := NewClient(address, username)
	if err := client.Connect(); err != nil {
		panic(err)
	}
	client.Run()
}
