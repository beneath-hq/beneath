package transpilers

const UserSchemaAvro = `` +
	`{"name":"user","type":"record","fields":[` +
	`{"name":"id","type":{"type":"string","logicalType":"uuid"}},` +
	`{"name":"password_hash","type":{"name":"bytes32","type":"fixed","size":32}},` +
	`{"name":"password_salt","type":"bytes32"},` +
	`{"name":"favorite_number","type":"int"},` +
	`{"name":"favorite_color","type":["null","string"]},` +
	`{"name":"home","type":{` +
	`"name":"coords","type":"record","fields":[` +
	`{"name":"latitude","type":"double"},` +
	`{"name":"longitude","type":"double"}` +
	`]}},` +
	`{"name":"work","type":"coords"}` +
	`]}`

const UserSchemaAvroWithDoc = `` +
	`{"name":"user","doc":"User test","type":"record","fields":[` +
	`{"name":"id","type":{"type":"string","logicalType":"uuid"}},` +
	`{"name":"password_hash","type":{"name":"bytes32","type":"fixed","size":32}},` +
	`{"name":"password_salt","type":"bytes32"},` +
	`{"name":"favorite_number","type":"int"},` +
	`{"name":"favorite_color","type":["null","string"]},` +
	`{"name":"home","type":{` +
	`"name":"coords","doc":"Coords test","type":"record","fields":[` +
	`{"name":"latitude","type":"double"},` +
	`{"name":"longitude","type":"double"}` +
	`]}},` +
	`{"name":"work","type":"coords"}` +
	`]}`
