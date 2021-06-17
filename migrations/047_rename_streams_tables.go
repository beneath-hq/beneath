package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE streams RENAME TO tables;
			ALTER TABLE tables RENAME COLUMN stream_id TO table_id;
			ALTER TABLE tables RENAME COLUMN primary_stream_instance_id TO primary_table_instance_id;
			ALTER TABLE tables RENAME CONSTRAINT streams_pkey TO tables_pkey;
			ALTER TABLE tables RENAME CONSTRAINT streams_primary_stream_instance_id_fkey TO tables_primary_table_instance_id_fkey;
			ALTER TABLE tables RENAME CONSTRAINT streams_project_id_fkey TO tables_project_id_fkey;
			ALTER INDEX streams_project_id_name_key RENAME TO tables_project_id_name_uidx;
			
			ALTER TABLE stream_indexes RENAME TO table_indexes;
			ALTER TABLE table_indexes RENAME COLUMN stream_index_id TO table_index_id;
			ALTER TABLE table_indexes RENAME COLUMN stream_id TO table_id;
			ALTER TABLE table_indexes RENAME CONSTRAINT stream_indexes_pkey TO table_indexes_pkey;
			ALTER TABLE table_indexes RENAME CONSTRAINT stream_indexes_stream_id_fkey TO table_indexes_table_id_fkey;
			ALTER INDEX stream_indexes_stream_id_idx RENAME TO table_indexes_table_id_idx;

			ALTER TABLE stream_instances RENAME TO table_instances;
			ALTER TABLE table_instances RENAME COLUMN stream_instance_id TO table_instance_id;
			ALTER TABLE table_instances RENAME COLUMN stream_id TO table_id;
			ALTER TABLE table_instances RENAME CONSTRAINT stream_instances_pkey TO table_instances_pkey;
			ALTER TABLE table_instances RENAME CONSTRAINT stream_instances_stream_id_fkey TO table_instances_table_id_fkey;
			ALTER INDEX stream_instances_stream_id_version_idx RENAME TO table_instances_table_id_version_uidx;

			ALTER TABLE permissions_services_streams RENAME TO permissions_services_tables;
			ALTER TABLE permissions_services_tables RENAME COLUMN stream_id TO table_id;
			ALTER TABLE permissions_services_tables RENAME CONSTRAINT permissions_services_streams_pkey TO permissions_services_tables_pkey;
			ALTER TABLE permissions_services_tables RENAME CONSTRAINT permissions_services_streams_service_id_fkey TO permissions_services_tables_service_id_fkey;
			ALTER TABLE permissions_services_tables RENAME CONSTRAINT permissions_services_streams_stream_id_fkey TO permissions_services_tables_table_id_fkey;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			ALTER TABLE permissions_services_tables RENAME CONSTRAINT permissions_services_tables_table_id_fkey TO permissions_services_streams_stream_id_fkey;
			ALTER TABLE permissions_services_tables RENAME CONSTRAINT permissions_services_tables_service_id_fkey TO permissions_services_streams_service_id_fkey;
			ALTER TABLE permissions_services_tables RENAME CONSTRAINT permissions_services_tables_pkey TO permissions_services_streams_pkey;
			ALTER TABLE permissions_services_tables RENAME COLUMN table_id TO stream_id;
			ALTER TABLE permissions_services_tables RENAME TO permissions_services_streams;

			ALTER INDEX table_instances_table_id_version_uidx RENAME TO stream_instances_stream_id_version_idx;
			ALTER TABLE table_instances RENAME CONSTRAINT table_instances_table_id_fkey TO stream_instances_stream_id_fkey;
			ALTER TABLE table_instances RENAME CONSTRAINT table_instances_pkey TO stream_instances_pkey;
			ALTER TABLE table_instances RENAME COLUMN table_id TO stream_id;
			ALTER TABLE table_instances RENAME COLUMN table_instance_id TO stream_instance_id;
			ALTER TABLE table_instances RENAME TO stream_instances;

			ALTER INDEX table_indexes_table_id_idx RENAME TO stream_indexes_stream_id_idx;
			ALTER TABLE table_indexes RENAME CONSTRAINT table_indexes_table_id_fkey TO stream_indexes_stream_id_fkey;
			ALTER TABLE table_indexes RENAME CONSTRAINT table_indexes_pkey TO stream_indexes_pkey;
			ALTER TABLE table_indexes RENAME COLUMN table_id TO stream_id;
			ALTER TABLE table_indexes RENAME COLUMN table_index_id TO stream_index_id;
			ALTER TABLE table_indexes RENAME TO stream_indexes;
			
			ALTER INDEX tables_project_id_name_uidx RENAME TO streams_project_id_name_key;
			ALTER TABLE tables RENAME CONSTRAINT tables_project_id_fkey TO streams_project_id_fkey;
			ALTER TABLE tables RENAME CONSTRAINT tables_primary_table_instance_id_fkey TO streams_primary_stream_instance_id_fkey;
			ALTER TABLE tables RENAME CONSTRAINT tables_pkey TO streams_pkey;
			ALTER TABLE tables RENAME COLUMN primary_table_instance_id TO primary_stream_instance_id;
			ALTER TABLE tables RENAME COLUMN table_id TO stream_id;
			ALTER TABLE tables RENAME TO streams;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}
