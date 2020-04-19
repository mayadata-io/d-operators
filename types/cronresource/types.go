/*
Copyright 2020 The MayaData Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package types

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metac "openebs.io/metac/apis/metacontroller/v1alpha1"
)

// ScheduleType is a typed definition to determine the type
// of a schedule
type ScheduleType string

const (
	// ScheduleTypeRepeat creates & deletes the resource
	// repeatedly based on schedule specifications
	ScheduleTypeRepeat ScheduleType = "Repeat"

	// ScheduleTypeHourly creates & deletes the resource
	// once every hour
	ScheduleTypeHourly ScheduleType = "Hourly"

	// ScheduleTypeDaily creates & deletes the resource
	// once everyday
	ScheduleTypeDaily ScheduleType = "Daily"

	// ScheduleTypeWeekly creates & deletes the resource
	// once every week
	ScheduleTypeWeekly ScheduleType = "Weekly"

	// ScheduleTypeOnce triggers creation & deletion of the
	// the target resource only once at a specific time in
	// the lifetime of this CronResource
	ScheduleTypeOnce ScheduleType = "Once"

	// ScheduleTypeImmediately triggers creation & deletion
	// of the the target resource immediately & then as
	// specified in the schedule specifications
	ScheduleTypeImmediately ScheduleType = "Immediately"

	// ScheduleTypeImmediatelyOnce triggers creation & deletion
	// of the the target resource immediately & only once in the
	// lifetime of this CronResource
	ScheduleTypeImmediatelyOnce ScheduleType = "ImmediatelyOnce"
)

// ScheduleDay is a typed definition to define supported
// days to schedule a resource
type ScheduleDay string

const (
	// ScheduleDayEveryday represents all days as supported
	// schedule days
	ScheduleDayEveryday ScheduleDay = "Everyday"

	// ScheduleDaySun represents Sunday as a supported schedule day
	ScheduleDaySun ScheduleDay = "Sun"

	// ScheduleDaySunday represents Sunday as a supported schedule day
	ScheduleDaySunday ScheduleDay = "Sunday"

	// ScheduleDayMon represents Monday as a supported schedule day
	ScheduleDayMon ScheduleDay = "Mon"

	// ScheduleDayMonday represents Monday as a supported schedule day
	ScheduleDayMonday ScheduleDay = "Monday"

	// ScheduleDayTue represents Tuesday as a supported schedule day
	ScheduleDayTue ScheduleDay = "Tue"

	// ScheduleDayTuesday represents Tuesday as a supported schedule day
	ScheduleDayTuesday ScheduleDay = "Tuesday"

	// ScheduleDayWed represents Wednesday as a supported schedule day
	ScheduleDayWed ScheduleDay = "Wed"

	// ScheduleDayWednesday represents Wednesday as a supported schedule day
	ScheduleDayWednesday ScheduleDay = "Wednesday"

	// ScheduleDayThu represents Thursday as a supported schedule day
	ScheduleDayThu ScheduleDay = "Thu"

	// ScheduleDayThursday represents Thursday as a supported schedule day
	ScheduleDayThursday ScheduleDay = "Thursday"

	// ScheduleDayFri represents Friday as a supported schedule day
	ScheduleDayFri ScheduleDay = "Fri"

	// ScheduleDayFriday represents Friday as a supported schedule day
	ScheduleDayFriday ScheduleDay = "Friday"

	// ScheduleDaySat represents Saturday as a supported schedule day
	ScheduleDaySat ScheduleDay = "Sat"

	// ScheduleDaySaturday represents Saturday as a supported schedule day
	ScheduleDaySaturday ScheduleDay = "Saturday"
)

// CronResource is a kubernetes custom resource that defines
// the specifications to create & delete any kubernetes resource
// at specified intervals
type CronResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec CronResourceSpec `json:"spec"`
}

// CronResourceSpec defines the configuration required
// to create & delete a kubernetes resource at specified intervals
type CronResourceSpec struct {
	// Supend indicates if all operations of this cron
	// schedule should be stopped. The relevant resources
	// created due to CronResource should be deleted.
	Suspend *bool `json:"suspend,omitempty"`

	// CompletionCriteria can be specified to mark the
	// current schedule as completed or not
	CompletionCriteria CronResourceCompletionCriteria `json:"completionCriteria,omitempty"`

	// Schedule to be followed for this CronResource
	Schedule CronResourceSchedule `json:"schedule,omitempty"`

	// Template has the desired state of the resource that
	// needs to be created or deleted as part of CronResource
	// schedule
	Template CronResourceTemplateReference `json:"template,omitempty"`
}

// CronResourceSchedule provides the schedule details that the
// entire CronResource reconciliation is based on
type CronResourceSchedule struct {
	// ScheduleType provides the frequency of this schedule
	//
	// Defaults to Hourly schedule
	ScheduleType ScheduleType `json:"scheduleType,omitempty"`

	// Time to start the schedule. In other words, post this
	// time schedule get activated.
	//
	// StartTime is optional
	StartTime metav1.Time `json:"startTime,omitempty"`

	// Time to end the schedule. Post this time resources created
	// due to this CronResource will get deleted.
	//
	// EndTime is optional
	EndTime metav1.Time `json:"endTime,omitempty"`

	// RepeatInterval indicates the interval post which a new
	// schedule instance gets triggered. In other words post
	// this interval the existing resource created due to the
	// schedule gets deleted & a new resource takes its place.
	// The repeat interval is ignored if existing resource is
	// not completed.
	//
	// RepeatInterval is used for ScheduleType **Repeat**
	//
	// RepeatInterval is optional & defaults to 1 hour if
	// ScheduleType is Repeat
	RepeatInterval time.Duration `json:"repeatInterval,omitempty"`

	// RepeatCount provides the maximum number of times a schedule
	// instance gets triggered
	//
	// If MaxCount is 5 and ScheduleType is Daily then, this
	// CronResource is supposed to run only for the next 5 days
	// with each new schedule occuring after 24 hours.
	//
	// RepeatCount is optional
	MaxCount *int `json:"maxCount,omitempty"`

	// Replicas / InstanceCount ?
	// cc @karthik can you clarify ?
	Replicas *int `json:"replicas,omitempty"`

	// EligibleDays provides the list of days that are eligible
	// to run this schedule. If there are any resources left due
	// to an earlier schedule on a non-eligible day, then this
	// resource gets deleted. No new resource will get created
	// on a non-eligible day.
	//
	// EligibleDays provides an extra filter to evaluate if
	// CronResource should be run or not.
	//
	// EligibleDays is optional & marks every day as eligible to
	// run this schedule
	EligibleDays []ScheduleDay `json:"eligibleDays,omitempty"`
}

// CronResourceCompletionCriteria is used to mark a CronResource
// schedule as completed or not
type CronResourceCompletionCriteria struct {
	// Select terms applied over the resource to mark the
	// current schedule as complete
	Selector metac.ResourceSelector `json:"selector,omitempty"`

	// Timeout indicates the time taken after which the
	// current schedule can be marked as completed
	Timeout time.Duration `json:"timeout,omitempty"`
}

// CronResourceTemplateReference provides the reference to the
// resource state or object or both. This is the resource that
// gets created & deleted as part of CronResource reconciliation.
type CronResourceTemplateReference struct {
	// Desired state that gets created & deleted as part of
	// CronResource reconciliation. This state can be in partial
	// form which in turn gets applied against the specifications
	// found in ObjectReference to arrive at the full resource
	// specs.
	//
	// NOTE:
	// - State is optional
	// - Both State & ObjectReference can not be empty
	// - Both State & ObjectReference can be set in a CronResource
	State map[string]interface{} `json:"state,omitempty"`

	// Resource instance has the resource state that gets created
	// & deleted as part of CronResource reconciliation
	//
	// NOTE:
	// - ObjectReference is optional
	// - Both ObjectReference & State can not be empty
	// - Both ObjectReference & State can be set in a CronResource
	ObjectReference CronResourceTemplateObjectReference `json:"objectReference,omitempty"`
}

// CronResourceTemplateObjectReference refers to the resource
// observed in the cluster that will be created & deleted as
// part of CronResource reconciliation
type CronResourceTemplateObjectReference struct {
	Name      string `json:"objectName"`
	Namespace string `json:"objectNamespace"`
}

// CronResourceTemplate is the placeholder for desired or default
// state of resource that gets created & deleted as part of
// CronResource reconciliation. This is a custom resource that can
// be applied separately against the kubernetes cluster.
type CronResourceTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec CronResourceTemplateSpec `json:"spec"`
}

// CronResourceTemplateSpec refers to the desired or default
// state of the resource that gets created & deleted as part
// of CronResource reconciliation
type CronResourceTemplateSpec struct {
	// State represents the desired state of the resource
	// that gets created & deleted as part of CronResource
	// reconciliation.
	State map[string]interface{} `json:"state,omitempty"`
}
