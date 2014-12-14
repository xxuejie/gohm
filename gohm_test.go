package gohm

import (
	`github.com/pote/redisurl`
	`testing`
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
