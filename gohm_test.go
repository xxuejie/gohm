package gohm

import (
	"github.com/pote/redisurl"
	"testing"
)

type user struct {
	ID    string `ohm:"id"`
	Name  string `ohm:"name"`
	Email string `ohm:"email index"`
	UUID  string `ohm:"uuid unique"`
	//Friends []user `ohm:"collection"`
}

func dbCleanup() {
	conn, _ := redisurl.Connect()

	conn.Do(`SCRIPT FLUSH`)
	conn.Do(`FLUSHDB`)
	conn.Close()
}

func TestSaveLoadsID(t *testing.T) {
	dbCleanup()
	defer dbCleanup()
	gohm, err := NewGohm()
	if err != nil {
		t.Error(err)
	}

	u := &user{
		Name:  `Marty`,
		Email: `marty@mcfly.com`,
	}

	err = gohm.Save(u)
	if err != nil {
		t.Error(err)
	}

	if u.ID != `1` {
		t.Errorf(`id is not set (expected "1", got "%v")`, u.ID)
	}
}

func TestFetchById(t *testing.T) {
	dbCleanup()
	defer dbCleanup()
	gohm, _ := NewGohm()
	gohm.Save(&user{
		Name:  `Marty`,
		Email: `marty@mcfly.com`,
	})

	u := &user{}
	found, err := gohm.FetchById(u, `1`)
	if err != nil {
		t.Error(err)
	}

	if !found {
		t.Errorf(`Found is not correct (expected true, was %v)`, found)
	}

	if u.ID != `1` {
		t.Errorf(`id not correctly set in model (expected "1", was "%v")`, u.ID)
	}

	if u.Name != `Marty` {
		t.Errorf(`incorrect Name set (expected "Marty", got "%v")`, u.Name)
	}
}

func TestFetchInvalidID(t *testing.T) {
	dbCleanup()
	defer dbCleanup()

	u := &user{}

	gohm, _ := NewGohm()

	found, err := gohm.FetchById(u, `1000000`)
	if err != nil {
		t.Error(err)
	}

	if found {
		t.Errorf(`Found is not correct (expected false, was %v)`, found)
	}
}

func TestDelete(t *testing.T) {
	dbCleanup()
	defer dbCleanup()
	gohm, _ := NewGohm()
	gohm.Save(&user{
		Name:  `Marty`,
		Email: `marty@mcfly.com`,
	})

	u := &user{}
	gohm.FetchById(u, `1`)

	err := gohm.Delete(u)
	if err != nil {
		t.Error(err)
	}

	// Make sure user is deleted indeed
	found, err := gohm.FetchById(u, `1`)
	if err != nil {
		t.Error(err)
	}

	if found {
		t.Errorf(`Found is not correct (expected false, was %v)`, found)
	}
}

func assertUserPresent(user user, users []user) bool {
	for i := range users {
		u := users[i]
		if user.ID == u.ID && user.Name == u.Name &&
			user.Email == u.Email && user.UUID == u.UUID {
			return true
		}
	}
	return false
}

func TestAll(t *testing.T) {
	dbCleanup()
	defer dbCleanup()
	gohm, err := NewGohm()
	if err != nil {
		t.Error(err)
	}

	u1 := &user{
		Name:  `Marty1`,
		Email: `marty1@mcfly.com`,
	}
	err = gohm.Save(u1)
	if err != nil {
		t.Error(err)
	}
	u2 := &user{
		Name:  `Marty2`,
		Email: `marty2@mcfly.com`,
	}
	err = gohm.Save(u2)
	if err != nil {
		t.Error(err)
	}

	var users []user
	err = gohm.All().Fetch(&users)
	if err != nil {
		t.Error(err)
	}

	if len(users) != 2 {
		t.Errorf(`Expected 2 users, but was %v`, len(users))
	}
	if !assertUserPresent(*u1, users) {
		t.Errorf(`Expected user "%v" to be presented but not`, u1.ID)
	}
	if !assertUserPresent(*u2, users) {
		t.Errorf(`Expected user "%v" to be presented but not`, u2.ID)
	}
}

func TestSingleReturnFromAll(t *testing.T) {
	dbCleanup()
	defer dbCleanup()
	gohm, err := NewGohm()
	if err != nil {
		t.Error(err)
	}

	expected := &user{
		Name:  `Marty1`,
		Email: `marty1@mcfly.com`,
	}
	err = gohm.Save(expected)
	if err != nil {
		t.Error(err)
	}

	var u user
	err = gohm.All().Fetch(&u)
	if err != nil {
		t.Error(err)
	}

	if !assertUserPresent(*expected, []user{u}) {
		t.Errorf(`Expected user "%v" to be presented but not`, expected.ID)
	}
}

func TestFilter(t *testing.T) {
	dbCleanup()
	defer dbCleanup()
	gohm, err := NewGohm()
	if err != nil {
		t.Error(err)
	}

	u1 := &user{
		Name:  `Marty1`,
		Email: `marty1@mcfly.com`,
	}
	err = gohm.Save(u1)
	if err != nil {
		t.Error(err)
	}
	u2 := &user{
		Name:  `Marty2`,
		Email: `marty2@mcfly.com`,
	}
	err = gohm.Save(u2)
	if err != nil {
		t.Error(err)
	}

	var users []user
	err = gohm.All().Find("email", "marty2@mcfly.com").Fetch(&users)
	if err != nil {
		t.Error(err)
	}

	if len(users) != 1 {
		t.Errorf(`Expected 1 user, but was %v`, len(users))
	}
	if !assertUserPresent(*u2, users) {
		t.Errorf(`Expected user "%v" to be presented but not`, u2.ID)
	}
}
