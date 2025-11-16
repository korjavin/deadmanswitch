package models

// ReminderUrgency represents the urgency level of a ping reminder
type ReminderUrgency string

const (
	// ReminderNormal indicates a routine check-in with > 24 hours until deadline
	ReminderNormal ReminderUrgency = "normal"
	// ReminderUrgent indicates an urgent check-in with 12-24 hours until deadline
	ReminderUrgent ReminderUrgency = "urgent"
	// ReminderFinalWarning indicates a final warning with < 12 hours until deadline
	ReminderFinalWarning ReminderUrgency = "final_warning"
)
