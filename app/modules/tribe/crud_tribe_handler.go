package tribe

import (
	"log"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/GORM-practice/app/helpers"
	"github.com/GORM-practice/app/models"
	"github.com/GORM-practice/app/modules/auth"
	"github.com/gorilla/mux"
)

// UintInSlice uint in slice
func UintInSlice(leads []models.TribeLeadAssign, targetUint uint64) bool {
	for _, lead := range leads {
		if uint64(lead.TribeID) == targetUint {
			return true
		}
	}
	return false
}

// CreateTribeHandler to handle createtribe
func (h *Handler) CreateTribeHandler(w http.ResponseWriter, r *http.Request) {
	// Get User ID
	_, role, err := auth.ExtractTokenUID(r)
	if err != nil {
		log.Println(err)
		helpers.SendError(w, "error uid extraction", http.StatusInternalServerError)
		return
	}
	if role < 1 {
		helpers.SendError(w, "super admin access only", http.StatusForbidden)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("[crud_tribe_handler.go][CreateTribeHandler][ReadBody]: %s\n", err)
		helpers.SendError(w, "error creating tribe", http.StatusBadRequest)
		return
	}

	tribe := tCreate{}
	if err = json.Unmarshal(body, &tribe); err != nil {
		fmt.Printf("[crud_tribe_handler.go][CreateTribeHandler][UnmarshalJSON]: %s\n", err)
		helpers.SendError(w, "error creating tribe", http.StatusBadRequest)
		return
	}

	if tribe.LeadID != 0 {
		var lead models.User
		if row := h.DB.Where("user_id = ?", tribe.LeadID).First(&lead); row.RowsAffected == 0 {
			helpers.SendError(w, "lead does not exist", http.StatusBadRequest)
			return
		}
	}

	tribeID, err := h.CreateTribe(tribe)
	if err != nil {
		fmt.Printf("[crud_tribe_handler.go][CreateTribeHandler][InsertTribe]: %s\n", err)
		helpers.SendError(w, "error creating tribe", http.StatusBadRequest)
		return
	}

	// helpers.SendOK(w, "tribe created")
	resp := map[string]interface{}{"status": true, "message": "create tribe success"}
	resp["tribe_id"] = tribeID
	write, _ := json.Marshal(resp)
	helpers.RenderJSON(w, write, http.StatusOK)
	return
}

//DeleteTribeHandler handle tribe deletion
func (h *Handler) DeleteTribeHandler(w http.ResponseWriter, r *http.Request) {
	_, role, err := auth.ExtractTokenUID(r)
	if err != nil {
		log.Println(err)
		helpers.SendError(w, "error uid extraction", http.StatusInternalServerError)
		return
	}
	if role < 1 {
		helpers.SendError(w, "super admin access only", http.StatusForbidden)
		return
	}

	params := mux.Vars(r)

	targetUint, err := strconv.ParseUint(params["tribe_id"], 10, 32)
	if err != nil {
		fmt.Printf("[crud_tribe_handler.go][DeleteTribeHandler][ParseUint]: %s\n", err)
		helpers.SendError(w, "error deleting tribe", http.StatusBadRequest)

		return
	}

	if err = h.DeleteTribe(uint(targetUint)); err != nil {
		fmt.Printf("[crud_tribe_handler.go][DeleteTribeHandler][DeleteTribe]: %s\n", err)
		helpers.SendError(w, "error deleting tribe", http.StatusBadRequest)
		return
	}

	helpers.SendOK(w, "tribe deleted")
	return
}

// UpdateTribeByID IMPROVE
func (h *Handler) UpdateTribeByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var tribe models.Tribe
	h.DB.First(&tribe, params["tribe_id"])

	uid, role, err := auth.ExtractTokenUID(r)
	if err != nil {
		helpers.SendError(w, "error UID extraction", http.StatusInternalServerError)
		return
	}

	if role < 1 && !UintInSlice(tribe.Leads, uid) { // Get user own key
		helpers.SendError(w, "You are not authorized for this request", http.StatusUnauthorized)
		return
	}

	//read edit info
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("[crud_tribe_handler.go][UpdateTribeByID][ReadBody]: %s\n", err)
		helpers.SendError(w, "Error when updating tribe", http.StatusBadRequest)
		return
	}

	updateTribe := models.Tribe{}
	if err = json.Unmarshal(body, &updateTribe); err != nil {
		fmt.Printf("[crud_tribe_handler.go][UpdateTribeByID][UnmarshalJSON]: %s\n", err)
		helpers.SendError(w, "Error when updating tribe", http.StatusBadRequest)
		return
	}

	UpdateValue(&updateTribe, &tribe)
	h.DB.Save(&tribe)
	helpers.SendOK(w, "Updated tribe")
}

//AddTribeLead IMPROVE
func (h *Handler) AddTribeLead(w http.ResponseWriter, r *http.Request) {
	//Superadmin handling
	_, role, err := auth.ExtractTokenUID(r)
	if err != nil {
		helpers.SendError(w, "error uid extraction", http.StatusInternalServerError)
		return
	}
	if role < 1 {
		helpers.SendError(w, "super admin access only", http.StatusForbidden)
		return
	}

	//get tribe uint64
	params := mux.Vars(r)
	tribeUint, err := strconv.ParseUint(params["tribe_id"], 10, 32)
	if err != nil {
		fmt.Printf("[crud_tribe_handler.go][AddTribeLead][ParseUint]: %s", err)
		helpers.SendError(w, "error assign user", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("[crud_tribe_handler.go][AddTribeLead][ReadBody]: %s\n", err)
		helpers.SendError(w, "error assign user", http.StatusBadRequest)
		return
	}

	var assign Assign
	//read body, get user id
	if err = json.Unmarshal(body, &assign); err != nil {
		fmt.Printf("[crud_tribe_handler.go][AddTribeLead][UnmarshalJSON]: %s\n", err)
		helpers.SendError(w, "error assign user", http.StatusBadRequest)
		return
	}

	var lead models.User
	if row := h.DB.First(&lead, assign.UID); row.RowsAffected == 0 {
		helpers.SendError(w, "user does not exist", http.StatusBadRequest)
		return
	}

	var tribe models.Tribe
	if row := h.DB.First(&tribe, uint(tribeUint)); row.RowsAffected == 0 {
		helpers.SendError(w, "tribe does not exist", http.StatusBadRequest)
		return
	}

	h.DB.Model(&tribe).Association("Leads").Append(models.TribeLeadAssign{LeadID: assign.UID, TribeID: uint(tribeUint)})
	h.DB.Model(&lead).Association("Tribes").Append(models.TribeAssign{UserID: assign.UID, TribeID: uint(tribeUint)})
	tribe.TotalMember = tribe.TotalMember + 1
	h.DB.Save(&tribe)
	helpers.SendOK(w, "Lead added")
	return
}

//RemoveTribeLead IMPROVE
func (h *Handler) RemoveTribeLead(w http.ResponseWriter, r *http.Request) {
	//Superadmin handling
	_, role, err := auth.ExtractTokenUID(r)

	if err != nil {
		helpers.SendError(w, "error uid extraction", http.StatusInternalServerError)
		return
	}
	if role < 1 {
		helpers.SendError(w, "super admin access only", http.StatusForbidden)
		return
	}

	//get tribe uint64
	params := mux.Vars(r)
	tribeUint, err := strconv.ParseUint(params["tribe_id"], 10, 32)
	if err != nil {
		fmt.Printf("[crud_tribe_handler.go][RemoveTribeLead][ParseUint]: %s", err)
		helpers.SendError(w, "error assign user", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("[crud_tribe_handler.go][RemoveTribeLead][ReadBody]: %s\n", err)
		helpers.SendError(w, "error assign user", http.StatusBadRequest)
		return
	}

	var assign Assign
	//read body, get user id
	if err = json.Unmarshal(body, &assign); err != nil {
		fmt.Printf("[crud_tribe_handler.go][RemoveTribeLead][UnmarshalJSON]: %s\n", err)
		helpers.SendError(w, "error assign user", http.StatusBadRequest)
		return
	}

	var lead models.User
	if row := h.DB.First(&lead, assign.UID); row.RowsAffected == 0 {
		helpers.SendError(w, "user does not exist", http.StatusBadRequest)
		return
	}

	var tribe models.Tribe
	if row := h.DB.First(&tribe, uint(tribeUint)); row.RowsAffected == 0 {
		helpers.SendError(w, "tribe does not exist", http.StatusBadRequest)
		return
	}

	if row := h.DB.Where("user_id = ? AND tribe_id = ?", assign.UID, tribeUint).Delete(models.TribeAssign{}); row.RowsAffected == 0 {
		helpers.SendError(w, "user is not assigned", http.StatusBadRequest)
		return
	}
	if row := h.DB.Where("lead_id = ? AND tribe_id = ?", assign.UID, tribeUint).Delete(models.TribeLeadAssign{}); row.RowsAffected == 0 {
		helpers.SendError(w, "user is not a lead", http.StatusBadRequest)
		return
	}
	tribe.TotalMember = tribe.TotalMember - 1
	h.DB.Save(&tribe)
	helpers.SendOK(w, "Lead removed")
	return
}

// GetTribeByID get tribe by id
func (h *Handler) GetTribeByID(w http.ResponseWriter, r *http.Request) {
	// Get User ID
	// _, role, err := auth.ExtractTokenUID(r)
	// if err != nil {
	// 	helpers.SendError(w, "error uid extraction", http.StatusInternalServerError)
	// 	return
	// }
	// if role < 1 {
	// 	helpers.SendError(w, "super admin access only", http.StatusForbidden)
	// 	return
	// }

	params := mux.Vars(r)
	var tribe models.Tribe
	h.DB.Preload("Members").Preload("Leads").Preload("Keys").First(&tribe, params["tribe_id"])
	write, _ := json.Marshal(&tribe)
	helpers.RenderJSON(w, write, http.StatusOK)
}

// TODO: GET USER BY EMAIL

// AssignUser assign user in tribe by lead
func (h *Handler) AssignUser(w http.ResponseWriter, r *http.Request) {

	//get tribe uint64
	params := mux.Vars(r)
	tribeUint, err := strconv.ParseUint(params["tribe_id"], 10, 32)
	if err != nil {
		fmt.Printf("[crud_tribe_handler.go][AssignUser][ParseUint]: %s", err)
		helpers.SendError(w, "error assign user", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("[crud_tribe_handler.go][AssignUser][ReadBody]: %s\n", err)
		helpers.SendError(w, "error assign user", http.StatusBadRequest)
		return
	}

	var assign Assign
	//read body, get user id
	if err = json.Unmarshal(body, &assign); err != nil {
		fmt.Printf("[crud_tribe_handler.go][AssignUser][UnmarshalJSON]: %s\n", err)
		helpers.SendError(w, "error assign user", http.StatusBadRequest)
		return
	}

	var user models.User
	if row := h.DB.First(&user, assign.UID); row.RowsAffected == 0 {
		helpers.SendError(w, "user does not exist", http.StatusBadRequest)
		return
	}

	var tribe models.Tribe
	if row := h.DB.Preload("Leads").First(&tribe, uint(tribeUint)); row.RowsAffected == 0 {
		helpers.SendError(w, "tribe does not exist", http.StatusBadRequest)
		return
	}

	// Get User ID
	uid, role, err := auth.ExtractTokenUID(r)
	if err != nil {
		helpers.SendError(w, "error uid extraction", http.StatusInternalServerError)
		return
	}

	if role < 1 && UintInSlice(tribe.Leads, uid) {
		helpers.SendError(w, "super admin access only", http.StatusForbidden)
		return
	}

	h.DB.Model(&tribe).Association("Members").Append(models.TribeAssign{UserID: assign.UID, TribeID: uint(tribeUint)})
	tribe.TotalMember = tribe.TotalMember + 1
	h.DB.Save(&tribe)
	helpers.SendOK(w, "user assigned")
	return
}

// RemoveAssign remove user from tribe by lead
func (h *Handler) RemoveAssign(w http.ResponseWriter, r *http.Request) {

	//get tribe uint64
	params := mux.Vars(r)
	tribeUint, err := strconv.ParseUint(params["tribe_id"], 10, 32)
	if err != nil {
		fmt.Printf("[crud_tribe_handler.go][RemoveAssign][ParseUint]: %s\n", err)
		helpers.SendError(w, "error remove user", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("[crud_tribe_handler.go][RemoveAssign][ReadBody]: %s\n", err)
		helpers.SendError(w, "error remove user", http.StatusBadRequest)
		return
	}

	var assign Assign
	//read body, get user id
	if err = json.Unmarshal(body, &assign); err != nil {
		fmt.Printf("[crud_tribe_handler.go][RemoveAssign][UnmarshalJSON]: %s\n", err)
		helpers.SendError(w, "error remove user", http.StatusBadRequest)
		return
	}

	// Get User ID
	uid, role, err := auth.ExtractTokenUID(r)
	if err != nil {
		helpers.SendError(w, "error uid extraction", http.StatusInternalServerError)
		return
	}

	var tribe models.Tribe
	h.DB.Preload("Leads").First(&tribe, uint(tribeUint))

	if role < 1 && UintInSlice(tribe.Leads, uid) {
		helpers.SendError(w, "tribe lead or super admin access only", http.StatusForbidden)
		return
	}

	if row := h.DB.Where("user_id = ? AND tribe_id = ?", assign.UID, tribeUint).Delete(models.TribeAssign{}); row.RowsAffected == 0 {
		helpers.SendError(w, "user is not assigned", http.StatusBadRequest)
		return
	}
	tribe.TotalMember = tribe.TotalMember - 1
	h.DB.Save(&tribe)
	helpers.SendOK(w, "removed user")
	return
}

// GetUserByTribeID returns as it says
func (h *Handler) GetUserByTribeID(w http.ResponseWriter, r *http.Request) {
	// Get User ID
	uid, role, err := auth.ExtractTokenUID(r)
	if err != nil {
		helpers.SendError(w, "error uid extraction", http.StatusInternalServerError)
		return
	}

	params := mux.Vars(r)
	var tribe models.Tribe
	h.DB.Preload("Members").Find(&tribe, params["tribe_id"])

	if role < 1 && UintInSlice(tribe.Leads, uid) {
		helpers.SendError(w, "tribe lead or super admin access only", http.StatusForbidden)
		return
	}

	var userIDs []uint
	for _, member := range tribe.Members {
		userIDs = append(userIDs, member.UserID)
	}

	var users []models.User
	for _, userID := range userIDs {
		var user models.User
		h.DB.First(&user, userID)
		users = append(users, user)
	}

	write, _ := json.Marshal(&users)
	helpers.RenderJSON(w, write, http.StatusOK)
}

// GetLeadByTribeID returns as it says
func (h *Handler) GetLeadByTribeID(w http.ResponseWriter, r *http.Request) {
	// Get User ID
	_, role, err := auth.ExtractTokenUID(r)
	if err != nil {
		helpers.SendError(w, "error uid extraction", http.StatusInternalServerError)
		return
	}

	params := mux.Vars(r)
	var tribe []models.TribeLeadAssign
	h.DB.Where("tribe_id = ? ", params["tribe_id"]).Find(&tribe)

	if role < 1 {
		helpers.SendError(w, "super admin access only", http.StatusForbidden)
		return
	}

	var leadIDs []uint
	for _, lead := range tribe {
		leadIDs = append(leadIDs, lead.LeadID)
	}

	var leads []models.User
	for _, leadID := range leadIDs {
		var lead models.User
		h.DB.First(&lead, leadID)
		leads = append(leads, lead)
	}

	write, _ := json.Marshal(&leads)
	helpers.RenderJSON(w, write, http.StatusOK)
}

// GetAllTribes returns all tribe
func (h *Handler) GetAllTribes(w http.ResponseWriter, r *http.Request) {
	// Get User ID
	_, role, err := auth.ExtractTokenUID(r)
	if err != nil {
		helpers.SendError(w, "error UID extraction", http.StatusInternalServerError)
		return
	}
	if role < 1 {
		helpers.SendError(w, "Request denied, superadmin only", http.StatusUnauthorized)
		return
	}

	var tribes []models.Tribe
	h.DB.Preload("Members").Preload("Leads").Preload("Keys").Order("tribe_id desc").Find(&tribes)
	write, _ := json.Marshal(&tribes)
	helpers.RenderJSON(w, write, http.StatusOK)
}

// GetUserNotLeadByTribeID returns user list that is not lead in the specified tribe
func (h *Handler) GetUserNotLeadByTribeID(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	// Get User ID
	_, role, err := auth.ExtractTokenUID(r)
	if err != nil {
		helpers.SendError(w, "error UID extraction", http.StatusInternalServerError)
		return
	}
	if role < 1 {
		helpers.SendError(w, "Request denied, superadmin only", http.StatusUnauthorized)
		return
	}

	var tribe models.Tribe
	if row := h.DB.First(&tribe, params["tribe_id"]).RowsAffected; row == 0 {
		helpers.SendError(w, "tribe does not exist", http.StatusBadRequest)
		return
	}

	var tla []models.TribeLeadAssign
	h.DB.Where("tribe_id = ?", params["tribe_id"]).Find(&tla)

	var leadsID []uint
	for _, v := range tla {
		leadsID = append(leadsID, v.LeadID)
	}

	var users []models.User
	h.DB.Not(leadsID).Find(&users, "role != ?", 1)

	write, _ := json.Marshal(&users)
	helpers.RenderJSON(w, write, http.StatusOK)
}
