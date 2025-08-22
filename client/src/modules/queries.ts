import { createProtobufRpcClient, QueryClient } from "@cosmjs/stargate"
import {VideoUpscalerLogs, VideoUpscalerTask, Worker} from '../types/generated/janction/videoUpscaler/v1/types'
import {QueryClientImpl, QueryGetVideoUpscalerLogsResponse, QueryGetWorkerResponse} from '../types/generated/janction/videoUpscaler/v1/query'
import {QueryGetVideoUpscalerTaskResponse} from '../../src/types/generated/janction/videoUpscaler/v1/query'

export interface VideoUpscalerExtension {
    readonly videoUpscaler: {
        readonly GetVideoUpscalerTask: (
            index: string
        ) => Promise<VideoUpscalerTask | undefined>;

        readonly GetVideoUpscalerLog: (
            threadId: string
        ) => Promise<VideoUpscalerLogs | undefined>;
        
        readonly GetWorker: (
            worker: string
        ) => Promise<Worker | undefined>;
    };
}

export function setupVideoUpscalerExtension(base: QueryClient): VideoUpscalerExtension {
    const rpc = createProtobufRpcClient(base);
    const queryService = new QueryClientImpl(rpc);

    return {
        videoUpscaler: {
            GetVideoUpscalerTask: async (index: string): Promise<VideoUpscalerTask | undefined> => {
                const response: QueryGetVideoUpscalerTaskResponse = await queryService.GetVideoUpscalerTask({
                    index: index,
                });
                return response.videoUpscalerTask;
            },

            GetVideoUpscalerLog: async (threadId: string): Promise<VideoUpscalerLogs | undefined> => {
                const response: QueryGetVideoUpscalerLogsResponse = await queryService.GetVideoUpscalerLogs({
                    threadId: threadId,
                });
                return response.videoUpscalerLogs;
            },

            GetWorker: async (worker: string): Promise<Worker | undefined> => {
                const response: QueryGetWorkerResponse = await queryService.GetWorker({
                    worker: worker,
                });
                return response.worker;
            },
        },
    };
}
