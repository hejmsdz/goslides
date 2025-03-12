package main

/*
	// OrgId    *uint
	// Org      Org
}

type Org struct {
	gorm.Model
	UUID uuid.UUID `gorm:"uniqueIndex"`
	Name string
}
func (o *Org) BeforeSave(tx *gorm.DB) (err error) {
	if o.UUID == uuid.Nil {
		o.UUID = uuid.New()
	}

	return nil
}


func (sdb SQLSongsDB) FilterSongs(query string, orgUUID string) []Song {
	querySlug := Slugify(query)

	var dbSongs []DbSong
	q := sdb.db.Select("uuid", "title", "subtitle", "number", "tags", "slug")
	if orgUUID == "" {
		q = q.Where("org_id IS NULL")
	} else {
		var org Org
		result := sdb.db.Where("uuid", orgUUID).Take(&org)
		if result.Error != nil {
			return []Song{}
		}

		q = q.Where("org_id IS NULL OR org_id = ?", org.ID)
	}

	// Where("org_id IS NULL OR org_id = ?").
	q.Where("slug LIKE ?", "%"+querySlug+"%").Order("title ASC").Find(&dbSongs)

func TestOrgSongsFiltering(t *testing.T) {
	sdb := SQLSongsDB{}
	sdb.Initialize("test.db")

	roch := Org{Name: "Schola DA św. Rocha"}
	zebrani := Org{Name: "Zebrani w Dnia Połowie"}
	sdb.db.Create([]*Org{&roch, &zebrani})

	witaj_pokarmie := DbSong{Title: "Witaj, pokarmie"}
	na_adwent := DbSong{Title: "Na Adwent", OrgId: &zebrani.ID}
	psalm_126 := DbSong{Title: "Psalm 126", OrgId: &roch.ID}
	sdb.db.Create([]*DbSong{&witaj_pokarmie, &na_adwent, &psalm_126})

	public_songs := sdb.FilterSongs("", "")
	if len(public_songs) != 1 || public_songs[0].Title != witaj_pokarmie.Title {
		t.Errorf("Expected to find 1 public song, got %+v", public_songs)
	}

	roch_songs := sdb.FilterSongs("", roch.UUID.String())
	if len(roch_songs) != 2 || roch_songs[0].Title != psalm_126.Title || roch_songs[1].Title != witaj_pokarmie.Title {
		t.Errorf("Expected to find 2 roch songs, got %+v", roch_songs)
	}

	zebrani_songs := sdb.FilterSongs("", zebrani.UUID.String())
	if len(zebrani_songs) != 2 || zebrani_songs[0].Title != na_adwent.Title || zebrani_songs[1].Title != witaj_pokarmie.Title {
		t.Errorf("Expected to find 2 zebrani songs, got %+v", zebrani_songs)
	}

	// if len(result) != 2 {
	// t.Errorf("Expected the line to be broken into 2 lines")
	// }
}
*/
