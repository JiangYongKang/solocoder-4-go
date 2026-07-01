package enrollment

import (
	"fmt"
	"time"
)

func (s *Store) generateID() string {
	s.idCounter++
	return fmt.Sprintf("%d", s.idCounter)
}

func (s *Store) getEnrollmentCount(courseID string) int {
	count := 0
	for _, e := range s.enrollments {
		if e.CourseID == courseID && e.Status == StatusEnrolled {
			count++
		}
	}
	return count
}

func (s *Store) isEnrolled(studentID, courseID string) bool {
	for _, e := range s.enrollments {
		if e.StudentID == studentID && e.CourseID == courseID && e.Status == StatusEnrolled {
			return true
		}
	}
	return false
}

func (s *Store) isInWaitlist(studentID, courseID string) bool {
	waitlist, exists := s.waitlists[courseID]
	if !exists {
		return false
	}
	for _, entry := range waitlist {
		if entry.StudentID == studentID {
			return true
		}
	}
	return false
}

func (s *Store) checkPrerequisites(student *Student, course *Course) bool {
	for _, prereqID := range course.Prerequisite {
		if !student.CompletedCourses[prereqID] {
			return false
		}
	}
	return true
}

func (s *Store) getEnrolledCredits(studentID string) int {
	total := 0
	for _, e := range s.enrollments {
		if e.StudentID == studentID && e.Status == StatusEnrolled {
			if course, exists := s.courses[e.CourseID]; exists {
				total += course.Credits
			}
		}
	}
	return total
}

func (s *Store) getCurrentEnrollment(studentID, courseID string) (*Enrollment, bool) {
	for _, e := range s.enrollments {
		if e.StudentID == studentID && e.CourseID == courseID && e.Status == StatusEnrolled {
			return e, true
		}
	}
	return nil, false
}

func (s *Store) removeFromWaitlist(courseID, studentID string) {
	waitlist, exists := s.waitlists[courseID]
	if !exists {
		return
	}
	for i, entry := range waitlist {
		if entry.StudentID == studentID {
			s.waitlists[courseID] = append(waitlist[:i], waitlist[i+1:]...)
			return
		}
	}
}

func (s *Store) processWaitlist(courseID string) (*Enrollment, error) {
	waitlist, exists := s.waitlists[courseID]
	if !exists || len(waitlist) == 0 {
		return nil, nil
	}

	course, _ := s.courses[courseID]
	for len(waitlist) > 0 {
		entry := waitlist[0]
		student, exists := s.students[entry.StudentID]
		if !exists {
			s.waitlists[courseID] = waitlist[1:]
			waitlist = s.waitlists[courseID]
			continue
		}

		if !s.checkPrerequisites(student, course) {
			s.waitlists[courseID] = waitlist[1:]
			waitlist = s.waitlists[courseID]
			continue
		}

		enrolledCredits := s.getEnrolledCredits(student.ID)
		if enrolledCredits+course.Credits > student.MaxCredits {
			s.waitlists[courseID] = waitlist[1:]
			waitlist = s.waitlists[courseID]
			continue
		}

		s.waitlists[courseID] = waitlist[1:]

		enrollment := &Enrollment{
			ID:         s.generateID(),
			StudentID:  student.ID,
			CourseID:   courseID,
			Status:     StatusEnrolled,
			EnrolledAt: time.Now(),
		}
		s.enrollments[enrollment.ID] = enrollment

		return enrollment, nil
	}

	return nil, nil
}

func (s *Store) Enroll(studentID, courseID string) (*Enrollment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.withinEnrollmentWindow() {
		return nil, ErrOutsideTimeWindow
	}

	student, exists := s.students[studentID]
	if !exists {
		return nil, ErrStudentNotFound
	}

	course, exists := s.courses[courseID]
	if !exists {
		return nil, ErrCourseNotFound
	}

	if s.isEnrolled(studentID, courseID) {
		return nil, ErrAlreadyEnrolled
	}

	if s.isInWaitlist(studentID, courseID) {
		return nil, ErrAlreadyInWaitlist
	}

	enrollmentCount := s.getEnrollmentCount(courseID)
	if enrollmentCount >= course.Capacity {
		waitlistEntry := &WaitlistEntry{
			StudentID: studentID,
			JoinedAt:  time.Now(),
		}
		s.waitlists[courseID] = append(s.waitlists[courseID], waitlistEntry)

		enrollment := &Enrollment{
			ID:         s.generateID(),
			StudentID:  studentID,
			CourseID:   courseID,
			Status:     StatusWaitlist,
			EnrolledAt: time.Now(),
		}
		s.enrollments[enrollment.ID] = enrollment

		return enrollment, ErrCourseFull
	}

	if !s.checkPrerequisites(student, course) {
		return nil, ErrPrerequisiteNotCompleted
	}

	enrolledCredits := s.getEnrolledCredits(studentID)
	if enrolledCredits+course.Credits > student.MaxCredits {
		return nil, ErrMaxCreditsExceeded
	}

	enrollment := &Enrollment{
		ID:         s.generateID(),
		StudentID:  studentID,
		CourseID:   courseID,
		Status:     StatusEnrolled,
		EnrolledAt: time.Now(),
	}
	s.enrollments[enrollment.ID] = enrollment

	return enrollment, nil
}

func (s *Store) Drop(studentID, courseID string) ([]*Enrollment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.withinEnrollmentWindow() {
		return nil, ErrOutsideTimeWindow
	}

	if _, exists := s.students[studentID]; !exists {
		return nil, ErrStudentNotFound
	}

	if _, exists := s.courses[courseID]; !exists {
		return nil, ErrCourseNotFound
	}

	if !s.isEnrolled(studentID, courseID) {
		if s.isInWaitlist(studentID, courseID) {
			s.removeFromWaitlist(courseID, studentID)
			return nil, nil
		}
		return nil, ErrNotEnrolled
	}

	enrollment, _ := s.getCurrentEnrollment(studentID, courseID)
	enrollment.Status = StatusDropped
	now := time.Now()
	enrollment.DroppedAt = &now

	waitlisted, err := s.processWaitlist(courseID)
	if err != nil {
		return nil, err
	}

	var promotedEnrollments []*Enrollment
	if waitlisted != nil {
		promotedEnrollments = append(promotedEnrollments, waitlisted)
	}

	return promotedEnrollments, nil
}

func (s *Store) GetEnrolledCourses(studentID string) ([]*Course, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.students[studentID]; !exists {
		return nil, ErrStudentNotFound
	}

	var courses []*Course
	for _, e := range s.enrollments {
		if e.StudentID == studentID && e.Status == StatusEnrolled {
			if course, exists := s.courses[e.CourseID]; exists {
				courses = append(courses, course)
			}
		}
	}
	return courses, nil
}

func (s *Store) GetWaitlistPosition(studentID, courseID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.students[studentID]; !exists {
		return 0, ErrStudentNotFound
	}

	if _, exists := s.courses[courseID]; !exists {
		return 0, ErrCourseNotFound
	}

	waitlist, exists := s.waitlists[courseID]
	if !exists {
		return 0, ErrNotInWaitlist
	}

	for i, entry := range waitlist {
		if entry.StudentID == studentID {
			return i + 1, nil
		}
	}

	return 0, ErrNotInWaitlist
}

func (s *Store) GetStudentCredits(studentID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.students[studentID]; !exists {
		return 0, ErrStudentNotFound
	}

	return s.getEnrolledCredits(studentID), nil
}

func (s *Store) GetAvailableSeats(courseID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	course, exists := s.courses[courseID]
	if !exists {
		return 0, ErrCourseNotFound
	}

	enrolled := s.getEnrollmentCount(courseID)
	return course.Capacity - enrolled, nil
}

func (s *Store) GetEnrollments() []*Enrollment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var enrollments []*Enrollment
	for _, e := range s.enrollments {
		enrollments = append(enrollments, e)
	}
	return enrollments
}
