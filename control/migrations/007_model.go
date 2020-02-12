package migrations

import (
	"github.com/beneath-core/control/entity"

	"github.com/go-pg/migrations/v7"
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
			ADD source_model_id UUID,
			ADD FOREIGN KEY (source_model_id) REFERENCES models (model_id) ON DELETE RESTRICT;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		// Stream.ModelID
		_, err = db.Exec(`
			ALTER TABLE streams DROP source_model_id;
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
