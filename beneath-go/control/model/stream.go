package model

import (
	"regexp"
	"time"

	uuid "github.com/satori/go.uuid"
	"gopkg.in/go-playground/validator.v9"
)

// constants
var (
	streamNameRegex *regexp.Regexp
)

// configure constants and validator
func init() {
	streamNameRegex = regexp.MustCompile("^[_a-z][_\\-a-z0-9]*$")
	GetValidator().RegisterStructValidation(streamValidation, Stream{})
}

// Stream represents a collection of data
type Stream struct {
	StreamID                uuid.UUID `sql:",pk,type:uuid"`
	Name                    string    `sql:",notnull",validate:"required,gte=1,lte=40"` // not unique because of (project_id, user_id) index
	Description             string    `validate:"omitempty,lte=255"`
	Schema                  string    `sql:",notnull",validate:"required"`
	AvroSchema              string    `sql:",type:json",validate:"required"`
	External                bool      `sql:",notnull"`
	Batch                   bool      `sql:",notnull"`
	Manual                  bool      `sql:",notnull"`
	ProjectID               uuid.UUID `sql:"on_delete:RESTRICT,notnull,type:uuid"`
	Project                 *Project
	StreamInstances         []*StreamInstance
	CurrentStreamInstanceID uuid.UUID `sql:"on_delete:SET NULL,type:uuid"`
	CurrentStreamInstance   *StreamInstance
	CreatedOn               time.Time `sql:",default:now()"`
	UpdatedOn               time.Time `sql:",default:now()"`
}

// custom stream validation
func streamValidation(sl validator.StructLevel) {
	s := sl.Current().Interface().(Stream)

	if !streamNameRegex.MatchString(s.Name) {
		sl.ReportError(s.Name, "Name", "", "alphanumericorunderscore", "")
	}
}

// FindOneStreamByNameAndProject finds a stream
func FindOneStreamByNameAndProject(name string, projectName string) *Stream {
	// TODO
	// return await getConnection()
	//   .createQueryBuilder(Stream, "stream")
	//   .innerJoinAndSelect("stream.project", "project")
	//   .where("stream.name = lower(:name)", { name })
	//   .andWhere("project.name = lower(:projectName)", { projectName })
	//   .getOne();
	return nil
}

/**
 * TODO: In the future
 * - belongs to model
 * - dependencies (models)
 * - define where to sink (i.e. do not presume pubsub, bq, bt)
 */
