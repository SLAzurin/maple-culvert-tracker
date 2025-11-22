import zod from 'zod';
import { buildServicesSchema } from './workflow-build-services-zod-schema';
import { writeFileSync } from 'node:fs';

const jsonSchema = zod.toJSONSchema(buildServicesSchema, { io: 'input' });
writeFileSync('./out/build-services.schema.json', JSON.stringify(jsonSchema));
