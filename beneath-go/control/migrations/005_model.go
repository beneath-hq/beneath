package migrations

import (
	"github.com/beneath-core/beneath-go/control/entity"

	"github.com/go-pg/migrations"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) (err error) {
		// Model
		err = db.Model(&entity.Model{}).CreateTable(defaultCreateOptions)
		if err != nil {
			return err
		}

		// (Project, name) unique index
		_, err = db.Exec(`
			CREATE UNIQUE INDEX models_project_id_name_key ON public.models USING btree (project_id, (lower(name)));
		`)
		if err != nil {
			return err
		}

		// StreamIntoModel
		err = db.Model(&entity.StreamIntoModel{}).CreateTable(defaultCreateOptions)
		if err != nil {
			return err
		}

		// Stream.ModelID
		_, err = db.Exec(`
			ALTER TABLE streams
			ADD model_id UUID,
			ADD FOREIGN KEY (model_id) REFERENCES models (model_id) ON DELETE RESTRICT;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Stream.ModelID
		_, err = db.Exec(`
			ALTER TABLE streams DROP model_id;
		`)
		if err != nil {
			return err
		}

		// StreamIntoModel
		err = db.Model(&entity.StreamIntoModel{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Model
		err = db.Model(&entity.Model{}).DropTable(defaultDropOptions)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}
