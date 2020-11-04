package migrations

import (
	"github.com/go-pg/migrations/v7"
)

func init() {
	Migrator.MustRegisterTx(func(db migrations.DB) (err error) {
		// BillingPlan
		_, err = db.Exec(`
			CREATE TABLE billing_plans(
				"billing_plan_id" uuid DEFAULT uuid_generate_v4(),
				"default" boolean NOT NULL DEFAULT false,
				"description" text,
				"created_on" timestamptz NOT NULL DEFAULT now(),
				"updated_on" timestamptz NOT NULL DEFAULT now(),
				"currency" text NOT NULL,
				"period" bigint NOT NULL,
				"seat_price_cents" integer NOT NULL,
				"seat_read_quota" bigint NOT NULL,
				"seat_write_quota" bigint NOT NULL,
				"base_read_quota" bigint NOT NULL,
				"base_write_quota" bigint NOT NULL,
				"read_overage_price_cents" integer NOT NULL,
				"write_overage_price_cents" integer NOT NULL,
				"personal" boolean NOT NULL,
				"private_projects" boolean NOT NULL,
				"available_in_ui" boolean NOT NULL,
				PRIMARY KEY ("billing_plan_id")
			)
		`)
		if err != nil {
			return err
		}

		// BilledResource
		_, err = db.Exec(`
			CREATE TABLE billed_resources(
				"billed_resource_id" uuid DEFAULT uuid_generate_v4(),
				"organization_id" uuid NOT NULL,
				"billing_time" timestamptz NOT NULL,
				"entity_id" uuid NOT NULL,
				"entity_name" text NOT NULL,
				"entity_kind" text NOT NULL,
				"start_time" timestamptz NOT NULL,
				"end_time" timestamptz NOT NULL,
				"product" text NOT NULL,
				"quantity" bigint NOT NULL,
				"total_price_cents" integer NOT NULL,
				"currency" text NOT NULL,
				"created_on" timestamptz NOT NULL DEFAULT now(),
				"updated_on" timestamptz NOT NULL DEFAULT now(),
				PRIMARY KEY ("billed_resource_id")
			)
		`)
		if err != nil {
			return err
		}

		// BillingInfo
		_, err = db.Exec(`
			CREATE TABLE billing_infos(
				billing_info_id       UUID DEFAULT uuid_generate_v4(),
				organization_id				UUID NOT NULL,
				billing_plan_id 			UUID NOT NULL,
				payments_driver 			TEXT NOT NULL,
				driver_payload 				JSONB,
				created_on 						timestamp with time zone DEFAULT now(),
				updated_on 						timestamp with time zone DEFAULT now(),
				PRIMARY KEY (billing_info_id),
				FOREIGN KEY (organization_id) REFERENCES organizations (organization_id) ON DELETE CASCADE,
				FOREIGN KEY (billing_plan_id) REFERENCES billing_plans (billing_plan_id) ON DELETE RESTRICT
			);
		`)
		if err != nil {
			return err
		}

		// only one default BillingPlan
		_, err = db.Exec(`
			CREATE UNIQUE INDEX ON billing_plans ("default") WHERE "default" = true;
		`)
		if err != nil {
			return err
		}

		// (billing_time, org_id, entity_id, product) unique index
		_, err = db.Exec(`
			CREATE UNIQUE INDEX billed_resources_billing_time_organization_id_entity_id_product_key
				ON public.billed_resources USING btree (billing_time, organization_id, entity_id, product);
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	}, func(db migrations.DB) (err error) {
		_, err = db.Exec(`
			DROP TABLE IF EXISTS billing_infos CASCADE;
			DROP TABLE IF EXISTS billed_resources CASCADE;
			DROP TABLE IF EXISTS billing_plans CASCADE;
		`)
		if err != nil {
			return err
		}

		// Done
		return nil
	})
}
