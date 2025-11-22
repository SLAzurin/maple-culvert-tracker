import { object, string, array } from 'zod';

export const buildServicesSchema = array(object({
    name: string(),
    dockerfile: string().optional().default("Dockerfile.goapps"),
    context: string().optional().default("."),
    target: string(),
    buildArgs: string().optional().default(''),
}));
