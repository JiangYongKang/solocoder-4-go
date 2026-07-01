package enrollment

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func setupTestStore() *Store {
	store := NewStore()

	store.AddStudent(&Student{
		ID:   "S001",
		Name: "Alice",
		Major: "Computer Science",
		CompletedCourses: map[string]bool{
			"C001": true,
		},
		MaxCredits: 20,
	})

	store.AddStudent(&Student{
		ID:   "S002",
		Name: "Bob",
		Major: "Computer Science",
		CompletedCourses: map[string]bool{
			"C001": true,
		},
		MaxCredits: 20,
	})

	store.AddStudent(&Student{
		ID:               "S003",
		Name:             "Charlie",
		Major:            "Computer Science",
		CompletedCourses: map[string]bool{},
		MaxCredits:       20,
	})

	store.AddCourse(&Course{
		ID:           "C001",
		Name:         "Introduction to Programming",
		Credits:      4,
		Capacity:     2,
		Prerequisite: []string{},
	})

	store.AddCourse(&Course{
		ID:           "C002",
		Name:         "Data Structures",
		Credits:      4,
		Capacity:     2,
		Prerequisite: []string{"C001"},
	})

	store.AddCourse(&Course{
		ID:           "C003",
		Name:         "Algorithms",
		Credits:      3,
		Capacity:     1,
		Prerequisite: []string{"C001", "C002"},
	})

	store.AddCourse(&Course{
		ID:           "C004",
		Name:         "Database Systems",
		Credits:      3,
		Capacity:     30,
		Prerequisite: []string{},
	})

	return store
}

func TestEnroll_Success(t *testing.T) {
	store := setupTestStore()

	enrollment, err := store.Enroll("S001", "C002")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if enrollment == nil {
		t.Fatal("Expected enrollment, got nil")
	}
	if enrollment.Status != StatusEnrolled {
		t.Errorf("Expected status ENROLLED, got %v", enrollment.Status)
	}
	if enrollment.StudentID != "S001" {
		t.Errorf("Expected student ID S001, got %v", enrollment.StudentID)
	}
	if enrollment.CourseID != "C002" {
		t.Errorf("Expected course ID C002, got %v", enrollment.CourseID)
	}
}

func TestEnroll_CourseFull(t *testing.T) {
	store := setupTestStore()

	_, err := store.Enroll("S001", "C001")
	if err != nil {
		t.Fatalf("S001 should enroll successfully, got %v", err)
	}

	_, err = store.Enroll("S002", "C001")
	if err != nil {
		t.Fatalf("S002 should enroll successfully, got %v", err)
	}

	enrollment, err := store.Enroll("S003", "C001")
	if err != ErrCourseFull {
		t.Fatalf("Expected ErrCourseFull, got %v", err)
	}
	if enrollment == nil {
		t.Fatal("Expected waitlist enrollment, got nil")
	}
	if enrollment.Status != StatusWaitlist {
		t.Errorf("Expected status WAITLIST, got %v", enrollment.Status)
	}

	position, err := store.GetWaitlistPosition("S003", "C001")
	if err != nil {
		t.Fatalf("Expected no error getting waitlist position, got %v", err)
	}
	if position != 1 {
		t.Errorf("Expected position 1, got %d", position)
	}
}

func TestEnroll_PrerequisiteNotCompleted(t *testing.T) {
	store := setupTestStore()

	_, err := store.Enroll("S003", "C002")
	if err != ErrPrerequisiteNotCompleted {
		t.Fatalf("Expected ErrPrerequisiteNotCompleted, got %v", err)
	}
}

func TestEnroll_MultiplePrerequisites(t *testing.T) {
	store := setupTestStore()

	_, err := store.Enroll("S001", "C003")
	if err != ErrPrerequisiteNotCompleted {
		t.Fatalf("Expected ErrPrerequisiteNotCompleted (missing C002), got %v", err)
	}

	store.students["S001"].CompletedCourses["C002"] = true

	_, err = store.Enroll("S001", "C003")
	if err != nil {
		t.Fatalf("Expected successful enrollment after completing prerequisites, got %v", err)
	}
}

func TestEnroll_AlreadyEnrolled(t *testing.T) {
	store := setupTestStore()

	_, err := store.Enroll("S001", "C002")
	if err != nil {
		t.Fatalf("First enrollment should succeed, got %v", err)
	}

	_, err = store.Enroll("S001", "C002")
	if err != ErrAlreadyEnrolled {
		t.Fatalf("Expected ErrAlreadyEnrolled, got %v", err)
	}
}

func TestEnroll_AlreadyInWaitlist(t *testing.T) {
	store := setupTestStore()

	_, err := store.Enroll("S001", "C001")
	if err != nil {
		t.Fatalf("S001 enrollment should succeed, got %v", err)
	}

	_, err = store.Enroll("S002", "C001")
	if err != nil {
		t.Fatalf("S002 enrollment should succeed, got %v", err)
	}

	_, err = store.Enroll("S003", "C001")
	if err != ErrCourseFull {
		t.Fatalf("S003 should go to waitlist, got %v", err)
	}

	_, err = store.Enroll("S003", "C001")
	if err != ErrAlreadyInWaitlist {
		t.Fatalf("Expected ErrAlreadyInWaitlist, got %v", err)
	}
}

func TestEnroll_StudentNotFound(t *testing.T) {
	store := setupTestStore()

	_, err := store.Enroll("S999", "C001")
	if err != ErrStudentNotFound {
		t.Fatalf("Expected ErrStudentNotFound, got %v", err)
	}
}

func TestEnroll_CourseNotFound(t *testing.T) {
	store := setupTestStore()

	_, err := store.Enroll("S001", "C999")
	if err != ErrCourseNotFound {
		t.Fatalf("Expected ErrCourseNotFound, got %v", err)
	}
}

func TestEnroll_MaxCreditsExceeded(t *testing.T) {
	store := setupTestStore()

	student := &Student{
		ID:               "S004",
		Name:             "David",
		Major:            "CS",
		CompletedCourses: map[string]bool{"C001": true},
		MaxCredits:       5,
	}
	store.AddStudent(student)

	_, err := store.Enroll("S004", "C002")
	if err != nil {
		t.Fatalf("C002 enrollment should succeed, got %v", err)
	}

	credits, err := store.GetStudentCredits("S004")
	if err != nil {
		t.Fatalf("GetStudentCredits failed: %v", err)
	}
	if credits != 4 {
		t.Errorf("Expected 4 credits, got %d", credits)
	}

	_, err = store.Enroll("S004", "C004")
	if err != ErrMaxCreditsExceeded {
		t.Fatalf("Expected ErrMaxCreditsExceeded (4+3=7 > 5), got %v", err)
	}
}

func TestEnroll_DefaultMaxCredits(t *testing.T) {
	store := NewStore()

	student := &Student{
		ID:               "S001",
		Name:             "Test",
		CompletedCourses: map[string]bool{},
	}
	store.AddStudent(student)

	if student.MaxCredits != 20 {
		t.Errorf("Expected default MaxCredits 20, got %d", student.MaxCredits)
	}
}

func TestDrop_Success(t *testing.T) {
	store := setupTestStore()

	_, err := store.Enroll("S001", "C002")
	if err != nil {
		t.Fatalf("Enrollment should succeed, got %v", err)
	}

	promoted, err := store.Drop("S001", "C002")
	if err != nil {
		t.Fatalf("Drop should succeed, got %v", err)
	}
	if len(promoted) != 0 {
		t.Errorf("Expected no promoted students, got %d", len(promoted))
	}

	courses, err := store.GetEnrolledCourses("S001")
	if err != nil {
		t.Fatalf("GetEnrolledCourses failed: %v", err)
	}
	if len(courses) != 0 {
		t.Errorf("Expected 0 enrolled courses, got %d", len(courses))
	}

	enrollments := store.GetEnrollments()
	var found bool
	for _, e := range enrollments {
		if e.StudentID == "S001" && e.CourseID == "C002" {
			found = true
			if e.Status != StatusDropped {
				t.Errorf("Expected status DROPPED, got %v", e.Status)
			}
			if e.DroppedAt == nil {
				t.Error("Expected DroppedAt to be set")
			}
		}
	}
	if !found {
		t.Error("Expected to find dropped enrollment record")
	}
}

func TestDrop_WithWaitlistPromotion(t *testing.T) {
	store := setupTestStore()

	_, err := store.Enroll("S001", "C001")
	if err != nil {
		t.Fatalf("S001 enrollment failed: %v", err)
	}

	_, err = store.Enroll("S002", "C001")
	if err != nil {
		t.Fatalf("S002 enrollment failed: %v", err)
	}

	_, err = store.Enroll("S003", "C001")
	if err != ErrCourseFull {
		t.Fatalf("S003 should be waitlisted, got %v", err)
	}

	seats, err := store.GetAvailableSeats("C001")
	if err != nil {
		t.Fatalf("GetAvailableSeats failed: %v", err)
	}
	if seats != 0 {
		t.Errorf("Expected 0 available seats, got %d", seats)
	}

	promoted, err := store.Drop("S001", "C001")
	if err != nil {
		t.Fatalf("Drop failed: %v", err)
	}

	if len(promoted) != 1 {
		t.Fatalf("Expected 1 promoted student, got %d", len(promoted))
	}
	if promoted[0].StudentID != "S003" {
		t.Errorf("Expected S003 to be promoted, got %v", promoted[0].StudentID)
	}
	if promoted[0].Status != StatusEnrolled {
		t.Errorf("Expected promoted student to be ENROLLED, got %v", promoted[0].Status)
	}

	seats, err = store.GetAvailableSeats("C001")
	if err != nil {
		t.Fatalf("GetAvailableSeats failed: %v", err)
	}
	if seats != 0 {
		t.Errorf("Expected 0 available seats after promotion, got %d", seats)
	}

	_, err = store.GetWaitlistPosition("S003", "C001")
	if err != ErrNotInWaitlist {
		t.Errorf("Expected ErrNotInWaitlist after promotion, got %v", err)
	}
}

func TestDrop_NotEnrolled(t *testing.T) {
	store := setupTestStore()

	_, err := store.Drop("S001", "C002")
	if err != ErrNotEnrolled {
		t.Fatalf("Expected ErrNotEnrolled, got %v", err)
	}
}

func TestDrop_RemoveFromWaitlist(t *testing.T) {
	store := setupTestStore()

	_, err := store.Enroll("S001", "C001")
	if err != nil {
		t.Fatalf("S001 enrollment failed: %v", err)
	}
	_, err = store.Enroll("S002", "C001")
	if err != nil {
		t.Fatalf("S002 enrollment failed: %v", err)
	}
	_, err = store.Enroll("S003", "C001")
	if err != ErrCourseFull {
		t.Fatalf("S003 should be waitlisted, got %v", err)
	}

	_, err = store.Drop("S003", "C001")
	if err != nil {
		t.Fatalf("Drop from waitlist should succeed, got %v", err)
	}

	_, err = store.GetWaitlistPosition("S003", "C001")
	if err != ErrNotInWaitlist {
		t.Errorf("Expected ErrNotInWaitlist, got %v", err)
	}
}

func TestDrop_StudentNotFound(t *testing.T) {
	store := setupTestStore()

	_, err := store.Drop("S999", "C001")
	if err != ErrStudentNotFound {
		t.Fatalf("Expected ErrStudentNotFound, got %v", err)
	}
}

func TestDrop_CourseNotFound(t *testing.T) {
	store := setupTestStore()

	_, err := store.Drop("S001", "C999")
	if err != ErrCourseNotFound {
		t.Fatalf("Expected ErrCourseNotFound, got %v", err)
	}
}

func TestTimeWindow_Setup(t *testing.T) {
	store := setupTestStore()

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now().Add(1 * time.Hour)

	err := store.SetEnrollmentWindow(start, end)
	if err != nil {
		t.Fatalf("SetEnrollmentWindow should succeed, got %v", err)
	}
}

func TestTimeWindow_InvalidRange(t *testing.T) {
	store := setupTestStore()

	start := time.Now().Add(1 * time.Hour)
	end := time.Now().Add(-1 * time.Hour)

	err := store.SetEnrollmentWindow(start, end)
	if err != ErrInvalidTimeRange {
		t.Fatalf("Expected ErrInvalidTimeRange, got %v", err)
	}
}

func TestTimeWindow_EnrollWithinWindow(t *testing.T) {
	store := setupTestStore()

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now().Add(1 * time.Hour)
	store.SetEnrollmentWindow(start, end)

	_, err := store.Enroll("S001", "C002")
	if err != nil {
		t.Fatalf("Enrollment within window should succeed, got %v", err)
	}
}

func TestTimeWindow_EnrollOutsideWindow(t *testing.T) {
	store := setupTestStore()

	start := time.Now().Add(1 * time.Hour)
	end := time.Now().Add(2 * time.Hour)
	store.SetEnrollmentWindow(start, end)

	_, err := store.Enroll("S001", "C002")
	if err != ErrOutsideTimeWindow {
		t.Fatalf("Expected ErrOutsideTimeWindow, got %v", err)
	}

	start = time.Now().Add(-2 * time.Hour)
	end = time.Now().Add(-1 * time.Hour)
	store.SetEnrollmentWindow(start, end)

	_, err = store.Enroll("S001", "C002")
	if err != ErrOutsideTimeWindow {
		t.Fatalf("Expected ErrOutsideTimeWindow (window closed), got %v", err)
	}
}

func TestTimeWindow_DropWithinWindow(t *testing.T) {
	store := setupTestStore()

	_, err := store.Enroll("S001", "C002")
	if err != nil {
		t.Fatalf("Enrollment failed: %v", err)
	}

	start := time.Now().Add(-1 * time.Hour)
	end := time.Now().Add(1 * time.Hour)
	store.SetEnrollmentWindow(start, end)

	_, err = store.Drop("S001", "C002")
	if err != nil {
		t.Fatalf("Drop within window should succeed, got %v", err)
	}
}

func TestTimeWindow_DropOutsideWindow(t *testing.T) {
	store := setupTestStore()

	_, err := store.Enroll("S001", "C002")
	if err != nil {
		t.Fatalf("Enrollment failed: %v", err)
	}

	start := time.Now().Add(1 * time.Hour)
	end := time.Now().Add(2 * time.Hour)
	store.SetEnrollmentWindow(start, end)

	_, err = store.Drop("S001", "C002")
	if err != ErrOutsideTimeWindow {
		t.Fatalf("Expected ErrOutsideTimeWindow, got %v", err)
	}
}

func TestGetEnrolledCourses(t *testing.T) {
	store := setupTestStore()

	_, err := store.Enroll("S001", "C002")
	if err != nil {
		t.Fatalf("Enrollment failed: %v", err)
	}

	_, err = store.Enroll("S001", "C004")
	if err != nil {
		t.Fatalf("Enrollment failed: %v", err)
	}

	courses, err := store.GetEnrolledCourses("S001")
	if err != nil {
		t.Fatalf("GetEnrolledCourses failed: %v", err)
	}
	if len(courses) != 2 {
		t.Errorf("Expected 2 courses, got %d", len(courses))
	}

	courseIDs := map[string]bool{}
	for _, c := range courses {
		courseIDs[c.ID] = true
	}
	if !courseIDs["C002"] || !courseIDs["C004"] {
		t.Errorf("Expected courses C002 and C004, got %v", courseIDs)
	}
}

func TestGetEnrolledCourses_StudentNotFound(t *testing.T) {
	store := setupTestStore()

	_, err := store.GetEnrolledCourses("S999")
	if err != ErrStudentNotFound {
		t.Fatalf("Expected ErrStudentNotFound, got %v", err)
	}
}

func TestGetStudentCredits(t *testing.T) {
	store := setupTestStore()

	_, err := store.Enroll("S001", "C002")
	if err != nil {
		t.Fatalf("Enrollment failed: %v", err)
	}

	_, err = store.Enroll("S001", "C004")
	if err != nil {
		t.Fatalf("Enrollment failed: %v", err)
	}

	credits, err := store.GetStudentCredits("S001")
	if err != nil {
		t.Fatalf("GetStudentCredits failed: %v", err)
	}
	if credits != 7 {
		t.Errorf("Expected 7 credits (4+3), got %d", credits)
	}
}

func TestGetStudentCredits_AfterDrop(t *testing.T) {
	store := setupTestStore()

	_, err := store.Enroll("S001", "C002")
	if err != nil {
		t.Fatalf("Enrollment failed: %v", err)
	}

	credits, err := store.GetStudentCredits("S001")
	if err != nil {
		t.Fatalf("GetStudentCredits failed: %v", err)
	}
	if credits != 4 {
		t.Errorf("Expected 4 credits, got %d", credits)
	}

	_, err = store.Drop("S001", "C002")
	if err != nil {
		t.Fatalf("Drop failed: %v", err)
	}

	credits, err = store.GetStudentCredits("S001")
	if err != nil {
		t.Fatalf("GetStudentCredits failed: %v", err)
	}
	if credits != 0 {
		t.Errorf("Expected 0 credits after drop, got %d", credits)
	}
}

func TestGetAvailableSeats(t *testing.T) {
	store := setupTestStore()

	seats, err := store.GetAvailableSeats("C002")
	if err != nil {
		t.Fatalf("GetAvailableSeats failed: %v", err)
	}
	if seats != 2 {
		t.Errorf("Expected 2 available seats, got %d", seats)
	}

	_, err = store.Enroll("S001", "C002")
	if err != nil {
		t.Fatalf("Enrollment failed: %v", err)
	}

	seats, err = store.GetAvailableSeats("C002")
	if err != nil {
		t.Fatalf("GetAvailableSeats failed: %v", err)
	}
	if seats != 1 {
		t.Errorf("Expected 1 available seat, got %d", seats)
	}
}

func TestGetAvailableSeats_CourseNotFound(t *testing.T) {
	store := setupTestStore()

	_, err := store.GetAvailableSeats("C999")
	if err != ErrCourseNotFound {
		t.Fatalf("Expected ErrCourseNotFound, got %v", err)
	}
}

func TestGetWaitlistPosition_NotInWaitlist(t *testing.T) {
	store := setupTestStore()

	_, err := store.GetWaitlistPosition("S001", "C001")
	if err != ErrNotInWaitlist {
		t.Fatalf("Expected ErrNotInWaitlist, got %v", err)
	}
}

func TestGetWaitlistPosition_StudentNotFound(t *testing.T) {
	store := setupTestStore()

	_, err := store.GetWaitlistPosition("S999", "C001")
	if err != ErrStudentNotFound {
		t.Fatalf("Expected ErrStudentNotFound, got %v", err)
	}
}

func TestGetWaitlistPosition_CourseNotFound(t *testing.T) {
	store := setupTestStore()

	_, err := store.GetWaitlistPosition("S001", "C999")
	if err != ErrCourseNotFound {
		t.Fatalf("Expected ErrCourseNotFound, got %v", err)
	}
}

func TestWaitlist_MultipleStudents(t *testing.T) {
	store := setupTestStore()

	store.AddStudent(&Student{
		ID:               "S004",
		Name:             "David",
		CompletedCourses: map[string]bool{"C001": true},
		MaxCredits:       20,
	})

	_, err := store.Enroll("S001", "C002")
	if err != nil {
		t.Fatalf("S001 enrollment failed: %v", err)
	}
	_, err = store.Enroll("S002", "C002")
	if err != nil {
		t.Fatalf("S002 enrollment failed: %v", err)
	}

	_, err = store.Enroll("S004", "C002")
	if err != ErrCourseFull {
		t.Fatalf("S004 should be waitlisted (pos 1), got %v", err)
	}

	_, err = store.Enroll("S001", "C001")
	if err != nil {
		t.Fatalf("S001 enrollment in C001 failed: %v", err)
	}
	_, err = store.Enroll("S002", "C001")
	if err != nil {
		t.Fatalf("S002 enrollment in C001 failed: %v", err)
	}
	_, err = store.Enroll("S003", "C001")
	if err != ErrCourseFull {
		t.Fatalf("S003 should be waitlisted for C001, got %v", err)
	}

	pos, err := store.GetWaitlistPosition("S004", "C002")
	if err != nil {
		t.Fatalf("GetWaitlistPosition failed: %v", err)
	}
	if pos != 1 {
		t.Errorf("Expected position 1 for S004, got %d", pos)
	}

	_, err = store.Drop("S001", "C002")
	if err != nil {
		t.Fatalf("Drop failed: %v", err)
	}

	pos, err = store.GetWaitlistPosition("S004", "C002")
	if err != ErrNotInWaitlist {
		t.Errorf("Expected S004 to be promoted, got %v", err)
	}

	courses, err := store.GetEnrolledCourses("S004")
	if err != nil {
		t.Fatalf("GetEnrolledCourses failed: %v", err)
	}
	found := false
	for _, c := range courses {
		if c.ID == "C002" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected S004 to be enrolled in C002 after promotion")
	}
}

func TestWaitlist_PromotionSkipsIneligibleStudents(t *testing.T) {
	store := setupTestStore()

	store.AddStudent(&Student{
		ID:               "S004",
		Name:             "David",
		CompletedCourses: map[string]bool{},
		MaxCredits:       20,
	})
	store.AddStudent(&Student{
		ID:               "S005",
		Name:             "Eve",
		CompletedCourses: map[string]bool{"C001": true},
		MaxCredits:       20,
	})

	_, err := store.Enroll("S001", "C002")
	if err != nil {
		t.Fatalf("S001 enrollment failed: %v", err)
	}
	_, err = store.Enroll("S002", "C002")
	if err != nil {
		t.Fatalf("S002 enrollment failed: %v", err)
	}

	_, err = store.Enroll("S004", "C002")
	if err != ErrCourseFull {
		t.Fatalf("S004 should be waitlisted, got %v", err)
	}
	_, err = store.Enroll("S005", "C002")
	if err != ErrCourseFull {
		t.Fatalf("S005 should be waitlisted, got %v", err)
	}

	pos4, _ := store.GetWaitlistPosition("S004", "C002")
	pos5, _ := store.GetWaitlistPosition("S005", "C002")
	if pos4 != 1 || pos5 != 2 {
		t.Errorf("Expected positions S004=1, S005=2, got %d and %d", pos4, pos5)
	}

	promoted, err := store.Drop("S001", "C002")
	if err != nil {
		t.Fatalf("Drop failed: %v", err)
	}

	if len(promoted) != 1 {
		t.Fatalf("Expected 1 promoted student, got %d", len(promoted))
	}
	if promoted[0].StudentID != "S005" {
		t.Errorf("Expected S005 (eligible) to be promoted, not S004 (ineligible), got %v", promoted[0].StudentID)
	}

	_, err = store.GetWaitlistPosition("S004", "C002")
	if err != ErrNotInWaitlist {
		t.Errorf("Expected S004 to be removed from waitlist (ineligible), got %v", err)
	}

	_, err = store.GetWaitlistPosition("S005", "C002")
	if err != ErrNotInWaitlist {
		t.Errorf("Expected S005 to be promoted, got %v", err)
	}
}

func TestWaitlist_PromotionSkipsCreditLimit(t *testing.T) {
	store := setupTestStore()

	store.AddStudent(&Student{
		ID:               "S004",
		Name:             "David",
		CompletedCourses: map[string]bool{"C001": true},
		MaxCredits:       3,
	})
	store.AddStudent(&Student{
		ID:               "S005",
		Name:             "Eve",
		CompletedCourses: map[string]bool{"C001": true},
		MaxCredits:       20,
	})

	_, err := store.Enroll("S001", "C002")
	if err != nil {
		t.Fatalf("S001 enrollment failed: %v", err)
	}
	_, err = store.Enroll("S002", "C002")
	if err != nil {
		t.Fatalf("S002 enrollment failed: %v", err)
	}

	_, err = store.Enroll("S004", "C002")
	if err != ErrCourseFull {
		t.Fatalf("S004 should be waitlisted, got %v", err)
	}
	_, err = store.Enroll("S005", "C002")
	if err != ErrCourseFull {
		t.Fatalf("S005 should be waitlisted, got %v", err)
	}

	promoted, err := store.Drop("S001", "C002")
	if err != nil {
		t.Fatalf("Drop failed: %v", err)
	}

	if len(promoted) != 1 {
		t.Fatalf("Expected 1 promoted student, got %d", len(promoted))
	}
	if promoted[0].StudentID != "S005" {
		t.Errorf("Expected S005 to be promoted (S004 has credit limit), got %v", promoted[0].StudentID)
	}
}

func TestConcurrentEnroll(t *testing.T) {
	store := NewStore()

	store.AddCourse(&Course{
		ID:       "C001",
		Name:     "Test Course",
		Credits:  3,
		Capacity: 5,
	})

	for i := 1; i <= 10; i++ {
		studentID := fmt.Sprintf("S%03d", i)
		store.AddStudent(&Student{
			ID:               studentID,
			Name:             fmt.Sprintf("Student %d", i),
			CompletedCourses: map[string]bool{},
			MaxCredits:       20,
		})
	}

	var wg sync.WaitGroup
	wg.Add(10)

	for i := 1; i <= 10; i++ {
		studentID := fmt.Sprintf("S%03d", i)
		go func(sid string) {
			defer wg.Done()
			store.Enroll(sid, "C001")
		}(studentID)
	}

	wg.Wait()

	enrolledCount := 0
	waitlistCount := 0
	for _, e := range store.GetEnrollments() {
		if e.CourseID == "C001" {
			if e.Status == StatusEnrolled {
				enrolledCount++
			} else if e.Status == StatusWaitlist {
				waitlistCount++
			}
		}
	}

	if enrolledCount > 5 {
		t.Errorf("Expected at most 5 enrolled students, got %d", enrolledCount)
	}
	if enrolledCount+waitlistCount != 10 {
		t.Errorf("Expected 10 total enrollments (enrolled + waitlist), got %d", enrolledCount+waitlistCount)
	}
}

func TestGetStudent_Exists(t *testing.T) {
	store := setupTestStore()

	student, exists := store.GetStudent("S001")
	if !exists {
		t.Fatal("Expected student to exist")
	}
	if student.ID != "S001" {
		t.Errorf("Expected student ID S001, got %v", student.ID)
	}
}

func TestGetStudent_NotExists(t *testing.T) {
	store := setupTestStore()

	_, exists := store.GetStudent("S999")
	if exists {
		t.Fatal("Expected student to not exist")
	}
}

func TestGetCourse_Exists(t *testing.T) {
	store := setupTestStore()

	course, exists := store.GetCourse("C001")
	if !exists {
		t.Fatal("Expected course to exist")
	}
	if course.ID != "C001" {
		t.Errorf("Expected course ID C001, got %v", course.ID)
	}
}

func TestGetCourse_NotExists(t *testing.T) {
	store := setupTestStore()

	_, exists := store.GetCourse("C999")
	if exists {
		t.Fatal("Expected course to not exist")
	}
}

func TestEnroll_NoPrerequisites(t *testing.T) {
	store := setupTestStore()

	_, err := store.Enroll("S003", "C004")
	if err != nil {
		t.Fatalf("Expected successful enrollment for course with no prerequisites, got %v", err)
	}
}

func TestDrop_EmptyWaitlist(t *testing.T) {
	store := setupTestStore()

	_, err := store.Enroll("S001", "C004")
	if err != nil {
		t.Fatalf("Enrollment failed: %v", err)
	}

	promoted, err := store.Drop("S001", "C004")
	if err != nil {
		t.Fatalf("Drop failed: %v", err)
	}
	if len(promoted) != 0 {
		t.Errorf("Expected no promotions with empty waitlist, got %d", len(promoted))
	}
}

func TestGetEnrollments_AllRecords(t *testing.T) {
	store := setupTestStore()

	_, err := store.Enroll("S001", "C002")
	if err != nil {
		t.Fatalf("Enrollment failed: %v", err)
	}
	_, err = store.Enroll("S002", "C002")
	if err != nil {
		t.Fatalf("Enrollment failed: %v", err)
	}
	_, err = store.Drop("S001", "C002")
	if err != nil {
		t.Fatalf("Drop failed: %v", err)
	}

	enrollments := store.GetEnrollments()
	if len(enrollments) != 2 {
		t.Errorf("Expected 2 enrollment records, got %d", len(enrollments))
	}
}

func TestCompletedCoursesNilInitialization(t *testing.T) {
	store := NewStore()

	student := &Student{
		ID:   "S001",
		Name: "Test",
	}
	store.AddStudent(student)

	if student.CompletedCourses == nil {
		t.Error("Expected CompletedCourses to be initialized, got nil")
	}
}

func TestNoTimeWindow_AlwaysAllowed(t *testing.T) {
	store := setupTestStore()

	_, err := store.Enroll("S001", "C002")
	if err != nil {
		t.Fatalf("Expected enrollment to succeed without time window, got %v", err)
	}

	_, err = store.Drop("S001", "C002")
	if err != nil {
		t.Fatalf("Expected drop to succeed without time window, got %v", err)
	}
}

func TestWaitlist_OrderPreserved(t *testing.T) {
	store := setupTestStore()

	store.AddStudent(&Student{
		ID:               "S004",
		Name:             "David",
		CompletedCourses: map[string]bool{"C001": true},
		MaxCredits:       20,
	})
	store.AddStudent(&Student{
		ID:               "S005",
		Name:             "Eve",
		CompletedCourses: map[string]bool{"C001": true},
		MaxCredits:       20,
	})
	store.AddStudent(&Student{
		ID:               "S006",
		Name:             "Frank",
		CompletedCourses: map[string]bool{"C001": true},
		MaxCredits:       20,
	})

	_, err := store.Enroll("S001", "C002")
	if err != nil {
		t.Fatalf("S001 enrollment failed: %v", err)
	}
	_, err = store.Enroll("S002", "C002")
	if err != nil {
		t.Fatalf("S002 enrollment failed: %v", err)
	}

	_, err = store.Enroll("S004", "C002")
	if err != ErrCourseFull {
		t.Fatalf("S004 should be waitlisted, got %v", err)
	}
	_, err = store.Enroll("S005", "C002")
	if err != ErrCourseFull {
		t.Fatalf("S005 should be waitlisted, got %v", err)
	}
	_, err = store.Enroll("S006", "C002")
	if err != ErrCourseFull {
		t.Fatalf("S006 should be waitlisted, got %v", err)
	}

	promoted, err := store.Drop("S001", "C002")
	if err != nil {
		t.Fatalf("First drop failed: %v", err)
	}
	if len(promoted) != 1 || promoted[0].StudentID != "S004" {
		t.Errorf("First promotion should be S004, got %v", promoted)
	}

	promoted, err = store.Drop("S002", "C002")
	if err != nil {
		t.Fatalf("Second drop failed: %v", err)
	}
	if len(promoted) != 1 || promoted[0].StudentID != "S005" {
		t.Errorf("Second promotion should be S005, got %v", promoted)
	}
}


