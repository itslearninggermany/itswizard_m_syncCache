package itswizard_m_syncCache

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/itslearninggermany/itswizard_m_basic"
	"github.com/itslearninggermany/itswizard_m_imses"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type SyncCache struct {
	PersonToDelete          []itswizard_m_basic.Person                    `json:"person_to_delete"`
	PersonToDeleteExist     bool                                          `json:"person_to_delete_exist"`
	PersonToImport          []itswizard_m_basic.Person                    `json:"person_to_import"`
	PersonToImportExist     bool                                          `json:"person_to_import_exist"`
	PersonToUpdate          []PersonUpdate                                `json:"person_to_update"`
	PersonToUpdateExist     bool                                          `json:"person_to_update_exist"`
	PersonsProblemsExist    bool                                          `json:"persons_problems_exist"`
	PersonsProblems         []PersonProblem                               `json:"persons_problems"`
	MsrProblemExist         bool                                          `json:"msr_problem_exist"`
	MsrProblem              []MsrProblem                                  `json:"msr_problem"`
	MsrToDeleteExist        bool                                          `json:"msr_to_delete_exist"`
	MsrToDelete             []itswizard_m_basic.MentorStudentRelationship `json:"msr_to_delete"`
	MsrToImportExist        bool                                          `json:"msr_to_import_exist"`
	MsrToImport             []itswizard_m_basic.MentorStudentRelationship `json:"msr_to_import"`
	SprProblemsExist        bool                                          `json:"spr_problems_exist"`
	SprProblem              []SprProblem                                  `json:"spr_problem"`
	SprToDeleteExist        bool                                          `json:"spr_to_delete_exist"`
	SprToDelete             []itswizard_m_basic.StudentParentRelationship `json:"spr_to_delete"`
	SprToImportExist        bool                                          `json:"spr_to_import_exist"`
	SprToImport             []itswizard_m_basic.StudentParentRelationship `json:"spr_to_import"`
	MembershipProblemsExist bool                                          `json:"membership_problems_exist"`
	MembershipProblems      []MembershipProblem                           `json:"membership_problems"`
	MembershipToImportExist bool                                          `json:"membership_to_import_exist"`
	MembershipToImport      []itswizard_m_basic.Membership                `json:"membership_to_import"`
	MembershipToDeleteExist bool                                          `json:"membership_to_delete_exist"`
	MembershipToDelete      []itswizard_m_basic.Membership                `json:"membership_to_delete"`
	GroupsToImportExist     bool                                          `json:"groups_to_import_exist"`
	GroupsToImport          []itswizard_m_basic.Group                     `json:"groups_to_import"`
	GroupsToDelete          []itswizard_m_basic.Group                     `json:"groups_to_delete"`
}

type PersonProblem struct {
	Person      itswizard_m_basic.Person `json:"person"`
	Information string                   `json:"information"`
}
type PersonUpdate struct {
	Person      itswizard_m_basic.Person `json:"person"`
	Information string                   `json:"information"`
}
type MembershipProblem struct {
	Information string                       `json:"information"`
	Membership  itswizard_m_basic.Membership `json:"membership"`
}
type MsrProblem struct {
	Problem string                                      `json:"problem"`
	Msr     itswizard_m_basic.MentorStudentRelationship `json:"msr"`
}
type SprProblem struct {
	Information string                                      `json:"information"`
	Spr         itswizard_m_basic.StudentParentRelationship `json:"spr"`
}

func (p *SyncCache) SaveCacheInJson(toFile bool) (output string, err error) {
	tmp, err := json.Marshal(&p)
	output = string(tmp)
	if toFile {
		err = ioutil.WriteFile(uuid.New().String()+".json", tmp, 666)
		if err != nil {
			return "", err
		}
	}
	return
}

func (p *SyncCache) SaveCacheInDatabase(organisationID uint, institutionID uint, db *gorm.DB) {
	tmp, err := json.Marshal(&p)
	if err != nil {
		log.Println(err)
		os.Exit(123)
	}
	err = db.Save(DbSyncCache{
		UserId:   0,
		Content:  string(tmp),
		Imported: false,
	}).Error
	if err != nil {
		log.Println(err)
		os.Exit(123)
	}
}

func GetCachefromJson(input string) (output SyncCache, err error) {
	err = json.Unmarshal([]byte(input), &output)
	return
}

func (p *SyncCache) Cache2Database(database *gorm.DB) (importSuccessfull bool, log string, err error) {
	var errorColector []string

	tx := database.Begin()
	// All New Persons
	for i := 0; i < len(p.PersonToImport); i++ {
		err = tx.Save(&p.PersonToImport[i]).Error
		log = log + fmt.Sprintln(p.PersonToImport[i].PersonSyncKey, " wurde hinzugefügt")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint("Person to import: ", p.PersonToImport[i].PersonSyncKey, err))
		}
	}

	for i := 0; i < len(p.PersonToUpdate); i++ {
		err = tx.Save(&p.PersonToUpdate[i].Person).Error
		log = log + fmt.Sprintln(p.PersonToUpdate[i].Person.PersonSyncKey, " wurde upgedated")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint("Person to update: ", p.PersonToUpdate[i].Person.PersonSyncKey, err))
		}
	}

	for i := 0; i < len(p.PersonToDelete); i++ {
		err = tx.Delete(&p.PersonToDelete[i]).Error
		log = log + fmt.Sprintln(p.PersonToDelete[i].PersonSyncKey, " wurde gelöscht")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint("Person to delete", p.PersonToDelete[i].PersonSyncKey, err))
		}
	}

	for i := 0; i < len(p.MsrToImport); i++ {
		err = tx.Save(&p.MsrToImport[i]).Error
		log = log + fmt.Sprintln(p.MsrToImport[i].StudentSyncPersonKey, "::", p.MsrToImport[i].MentorSyncPersonKey, " wurde hinzugefügt")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint("Msr to import: ", p.MsrToImport[i], err))
		}
	}

	for i := 0; i < len(p.MsrToDelete); i++ {
		err = tx.Delete(&p.MsrToDelete[i]).Error
		log = log + fmt.Sprintln(p.MsrToDelete[i].StudentSyncPersonKey, "::", p.MsrToDelete[i].MentorSyncPersonKey, " wurde gelöscht")

		if err != nil {
			errorColector = append(errorColector, fmt.Sprint("MsrToDelete: ", p.MsrToDelete[i], err))
		}
	}

	for i := 0; i < len(p.SprToImport); i++ {
		err = tx.Save(&p.SprToImport[i]).Error
		log = log + fmt.Sprintln(p.SprToImport[i].StudentSyncPersonKey, "::", p.SprToImport[i].ParentSyncPersonKey, " wurde hinzugefügt")

		if err != nil {
			errorColector = append(errorColector, fmt.Sprint("SPR to import", p.SprToImport[i], err))
		}
	}

	for i := 0; i < len(p.SprToDelete); i++ {
		err = tx.Delete(&p.SprToDelete[i]).Error
		log = log + fmt.Sprintln(p.SprToDelete[i].StudentSyncPersonKey, "::", p.SprToDelete[i].ParentSyncPersonKey, " wurde gelöscht")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint("SPR to Delete: ", p.SprToDelete[i], err))
		}
	}

	for i := 0; i < len(p.GroupsToImport); i++ {
		err = tx.Save(&p.GroupsToImport[i]).Error
		log = log + fmt.Sprintln(p.GroupsToImport[i].Name, " wurde hinzugefügt")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint("Groups to import: ", p.GroupsToImport[i], err))
		}
	}

	for i := 0; i < len(p.GroupsToDelete); i++ {
		err = tx.Delete(&p.GroupsToDelete[i]).Error
		log = log + fmt.Sprintln(p.GroupsToDelete[i].Name, " wurde entfernt")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint("Groups to delete: ", p.GroupsToDelete[i], err))
		}
	}

	for i := 0; i < len(p.MembershipToImport); i++ {
		err = tx.Save(&p.MembershipToImport[i]).Error
		log = log + fmt.Sprintln(p.MembershipToImport[i].PersonSyncKey, "::", p.MembershipToImport[i].GroupSyncKey, " wurde hinzugefügt")
		if err != nil {
			errorColector = append(errorColector, fmt.Sprint("Membership to import: ", p.MembershipToImport[i], err))
		}
	}

	var memberships []itswizard_m_basic.Membership
	if len(p.MembershipToDelete) > 0 {
		database.Where("organisation15 = ?", p.MembershipToDelete[0].Organisation15).Find(&memberships)
	}

	for i := 0; i < len(p.MembershipToDelete); i++ {
		for _, mem := range memberships {
			if mem.GroupSyncKey == p.MembershipToDelete[i].GroupSyncKey {
				if mem.PersonSyncKey == p.MembershipToDelete[i].PersonSyncKey {
					err = tx.Delete(&mem).Error
					log = log + fmt.Sprintln(p.MembershipToDelete[i].PersonSyncKey, "::", p.MembershipToDelete[i].GroupSyncKey, " wurde gelöscht")
					if err != nil {
						errorColector = append(errorColector, fmt.Sprint("Membership to delete: ", p.MembershipToDelete[i], err))
					}
					break
				}
			}
		}
	}

	if len(errorColector) > 0 {
		importSuccessfull = false
		err = errors.New(strings.Join(errorColector, ""))
		tx.Rollback()
	} else {
		importSuccessfull = true
		tx.Commit()
	}
	return
}

type DbSyncCache struct {
	gorm.Model
	UserId   uint
	Content  string `gorm:"type:MEDIUMTEXT"`
	Imported bool
}

func (p *SyncCache) Cache2ItslearningOverImses(institutionID, organisationID uint, rootGroup, username, passwort, endpoint string, databse *gorm.DB) {
	itsl := itswizard_m_imses.NewImsesService(itswizard_m_imses.NewImsesServiceInput{
		Username: username,
		Password: passwort,
		Url:      endpoint,
	})

	for _, person := range p.PersonToImport {
		resp, _ := itsl.CreatePerson(itswizard_m_basic.DbPerson15{
			SyncPersonKey: person.PersonSyncKey,
			FirstName:     person.FirstName,
			LastName:      person.LastName,
			Username:      person.Username,
			Profile:       person.Profile,
			Password:      person.Password,
			Email:         person.Email,
			Phone:         person.Phone,
			Mobile:        person.Mobile,
			Street1:       person.Street1,
			Street2:       person.Street2,
			Postcode:      person.Postcode,
			City:          person.City,
		})
		databse.Save(&itswizard_m_basic.ChangeLog{
			UserOrGroup:      person.PersonSyncKey,
			NewPerson:        true,
			DeltePerson:      false,
			UdpatePerson:     false,
			GroupImport:      false,
			GroupDelete:      false,
			MembershipImport: false,
			MembershipDelete: false,
			PSR:              false,
			Error:            resp,
			OrganisationId:   organisationID,
			InstitutionID:    institutionID,
		})

		profile := "Guest"
		if person.Profile == "Staff" {
			profile = "Instructor"
		}
		if person.Profile == "Student" {
			profile = "Learner"
		}
		resp, _ = itsl.CreateMembership(rootGroup, person.PersonSyncKey, profile)
		databse.Save(&itswizard_m_basic.ChangeLog{
			UserOrGroup:      person.PersonSyncKey + "++" + rootGroup,
			NewPerson:        false,
			DeltePerson:      false,
			UdpatePerson:     false,
			GroupImport:      false,
			GroupDelete:      false,
			MembershipImport: true,
			MembershipDelete: false,
			PSR:              false,
			Error:            resp,
			OrganisationId:   organisationID,
			InstitutionID:    institutionID,
		})
	}

	for _, person := range p.PersonToUpdate {
		resp, _ := itsl.CreatePerson(itswizard_m_basic.DbPerson15{
			SyncPersonKey: person.Person.PersonSyncKey,
			FirstName:     person.Person.FirstName,
			LastName:      person.Person.LastName,
			Username:      person.Person.Username,
			Profile:       person.Person.Profile,
			Password:      person.Person.Password,
			Email:         person.Person.Email,
			Phone:         person.Person.Phone,
			Mobile:        person.Person.Mobile,
			Street1:       person.Person.Street1,
			Street2:       person.Person.Street2,
			Postcode:      person.Person.Postcode,
			City:          person.Person.City,
		})

		databse.Save(&itswizard_m_basic.ChangeLog{
			UserOrGroup:      person.Person.PersonSyncKey + "++" + rootGroup,
			NewPerson:        false,
			DeltePerson:      false,
			UdpatePerson:     false,
			GroupImport:      false,
			GroupDelete:      false,
			MembershipImport: true,
			MembershipDelete: false,
			PSR:              false,
			Error:            resp,
			OrganisationId:   organisationID,
			InstitutionID:    institutionID,
		})
	}

	for _, person := range p.PersonToDelete {
		resp, _ := itsl.DeletePerson(person.PersonSyncKey)
		databse.Save(&itswizard_m_basic.ChangeLog{
			UserOrGroup:      person.PersonSyncKey + "++" + rootGroup,
			NewPerson:        false,
			DeltePerson:      true,
			UdpatePerson:     false,
			GroupImport:      false,
			GroupDelete:      false,
			MembershipImport: false,
			MembershipDelete: false,
			PSR:              false,
			Error:            resp,
			OrganisationId:   organisationID,
			InstitutionID:    institutionID,
		})

	}

	kursgruppeErstellt := false
	for _, group := range p.GroupsToImport {
		if group.IsCourse {
			if !kursgruppeErstellt {
				resp, _ := itsl.CreateGroup(itswizard_m_basic.DbGroup15{
					ID:            rootGroup + "kursGruppe",
					SyncID:        rootGroup + "kursGruppe",
					Name:          "Kurse",
					ParentGroupID: rootGroup,
					Level:         1,
					IsCourse:      false,
				}, false)
				databse.Save(&itswizard_m_basic.ChangeLog{
					UserOrGroup:      rootGroup + "kursGruppe",
					NewPerson:        false,
					DeltePerson:      false,
					UdpatePerson:     false,
					GroupImport:      true,
					GroupDelete:      false,
					MembershipImport: false,
					MembershipDelete: false,
					PSR:              false,
					Error:            resp,
					OrganisationId:   organisationID,
					InstitutionID:    institutionID,
				})
				kursgruppeErstellt = true
			}
			resp, _ := itsl.CreateCourse(itswizard_m_basic.DbGroup15{
				SyncID:        group.GroupSyncKey,
				Name:          group.Name,
				ParentGroupID: rootGroup + "kursGruppe",
				Level:         2,
				IsCourse:      true,
			})
			databse.Save(&itswizard_m_basic.ChangeLog{
				UserOrGroup:      group.GroupSyncKey,
				CreateCourse:     true,
				NewPerson:        false,
				DeltePerson:      false,
				UdpatePerson:     false,
				GroupImport:      false,
				GroupDelete:      false,
				MembershipImport: false,
				MembershipDelete: false,
				PSR:              false,
				Error:            resp,
				OrganisationId:   organisationID,
				InstitutionID:    institutionID,
			})
		}
		if !group.IsCourse {
			resp, _ := itsl.CreateGroup(itswizard_m_basic.DbGroup15{
				SyncID:        group.GroupSyncKey,
				Name:          group.Name,
				ParentGroupID: group.ParentGroupID,
			}, false)
			databse.Save(&itswizard_m_basic.ChangeLog{
				UserOrGroup:      group.GroupSyncKey,
				NewPerson:        false,
				DeltePerson:      false,
				UdpatePerson:     false,
				CreateCourse:     false,
				GroupImport:      true,
				GroupDelete:      false,
				MembershipImport: false,
				MembershipDelete: false,
				PSR:              false,
				Error:            resp,
				OrganisationId:   organisationID,
				InstitutionID:    institutionID,
			})
		}
	}

	for _, membership := range p.MembershipToImport {
		resp, _ := itsl.CreateMembership(membership.GroupSyncKey, membership.PersonSyncKey, membership.Profile)
		databse.Save(&itswizard_m_basic.ChangeLog{
			Model:            gorm.Model{},
			UserOrGroup:      membership.PersonSyncKey + "++" + membership.GroupSyncKey,
			NewPerson:        false,
			DeltePerson:      false,
			UdpatePerson:     false,
			GroupImport:      false,
			GroupDelete:      false,
			MembershipImport: true,
			MembershipDelete: false,
			PSR:              false,
			Error:            resp,
			OrganisationId:   organisationID,
			InstitutionID:    institutionID,
		})
	}

	for i := 0; i < len(p.MembershipToDelete); i++ {
		membershipsFromPerson := itsl.ReadMembershipsForPerson(p.MembershipToDelete[i].PersonSyncKey)
		for _, mem := range membershipsFromPerson {
			if mem.GroupID == p.MembershipToDelete[i].GroupSyncKey {
				resp, _ := itsl.DeleteMembership(mem.ID)
				databse.Save(&itswizard_m_basic.ChangeLog{
					UserOrGroup:      p.MembershipToImport[i].PersonSyncKey + "++" + p.MembershipToImport[i].GroupSyncKey,
					NewPerson:        false,
					DeltePerson:      false,
					UdpatePerson:     false,
					GroupImport:      false,
					GroupDelete:      false,
					MembershipImport: false,
					MembershipDelete: true,
					PSR:              false,
					Error:            resp,
					OrganisationId:   organisationID,
					InstitutionID:    institutionID,
				})
			}
		}
	}
}
