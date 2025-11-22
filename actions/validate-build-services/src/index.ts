import * as core from '@actions/core';
import { readFileSync } from 'node:fs'
import { buildServicesSchema } from './workflow-build-services-zod-schema';
import { ZodError } from 'zod';

void (async () => {
    try {
        const data = readFileSync('workflow-build-services.json', 'utf-8');
        const parsed = JSON.parse(data);

        const zodparsed = buildServicesSchema.parse(parsed);

        core.info('Build services configuration is valid.');
        core.setOutput("build-services", zodparsed);
    } catch (e) {
        core.error('Build services configuration is invalid.');
        if (e instanceof ZodError) {
            for (const issue of e.issues) {
                core.error(`- ${issue.path.join('.')} : ${issue.message}`);
            }
            core.setFailed('Validation failed due to schema errors.');
        } else if (e instanceof Error) {
            core.setFailed(e);
        } else {
            core.setFailed('An unknown error occurred.');
        }
    }
})();
