import { QueryClient, StargateClient, StargateClientOptions } from "@cosmjs/stargate"
import { Tendermint34Client } from "@cosmjs/tendermint-rpc"
import { VideoUpscalerExtension, setupVideoUpscalerExtension } from "./modules/queries"

export class VideoUpscalerStargateClient extends StargateClient {
    public readonly videoUpscalerQueryClient: VideoUpscalerExtension | undefined

    public static async connect(
        endpoint: string,
        options?: StargateClientOptions,
    ): Promise<VideoUpscalerStargateClient> {
        const tmClient = await Tendermint34Client.connect(endpoint)
        return new VideoUpscalerStargateClient(tmClient, options)
    }

    protected constructor(tmClient: Tendermint34Client | undefined, options: StargateClientOptions = {}) {
        super(tmClient, options)
        if (tmClient) {
            this.videoUpscalerQueryClient = QueryClient.withExtensions(tmClient, setupVideoUpscalerExtension)
        }
    }
}