import { GeneratedType, OfflineSigner, Registry } from "@cosmjs/proto-signing"
import {
    defaultRegistryTypes,
    DeliverTxResponse,
    QueryClient,
    SigningStargateClient,
    SigningStargateClientOptions,
    StdFee,
} from "@cosmjs/stargate"
import { Tendermint34Client } from "@cosmjs/tendermint-rpc"
import Long from "long"
import { VideoUpscalerExtension, setupVideoUpscalerExtension } from "./modules/queries"
import {
    videoUpscalerTypes,
    MsgCreateVideoUpscalerTaskEncodeObject,
    typeUrlMsgCreateVideoUpscalerTask,
} from "./modules/messages"

export const videoUpscalerDefaultRegistryTypes: ReadonlyArray<[string, GeneratedType]> = [
    ...defaultRegistryTypes,
    ...videoUpscalerTypes,
]

function createDefaultRegistry(): Registry {
    const registry = new Registry(defaultRegistryTypes)
    registry.register(videoUpscalerTypes[0][0],videoUpscalerTypes[0][1]  );
    registry.register(videoUpscalerTypes[1][0],videoUpscalerTypes[1][1]  );
    return registry
}

export class VideoUpscalerSigningStargateClient extends SigningStargateClient {
    public readonly checkersQueryClient: VideoUpscalerExtension | undefined

    public static async connectWithSigner(
        endpoint: string,
        signer: OfflineSigner,
        options: SigningStargateClientOptions = {},
    ): Promise<VideoUpscalerSigningStargateClient> {
        const tmClient = await Tendermint34Client.connect(endpoint)
        return new VideoUpscalerSigningStargateClient(tmClient, signer, {
            registry: createDefaultRegistry(),
            ...options,
        })
    }

    protected constructor(
        tmClient: Tendermint34Client | undefined,
        signer: OfflineSigner,
        options: SigningStargateClientOptions,
    ) {
        super(tmClient, signer, options)
        if (tmClient) {
            this.checkersQueryClient = QueryClient.withExtensions(tmClient, setupVideoUpscalerExtension)
        }
    }

    public async createVideoUpscalerTask(
        creator: string,
        cid: string,
        startFrame: number,
        endFrame: number,
        threads: number,
        reward: Long,
        fee: number | StdFee | "auto"
    ): Promise<DeliverTxResponse> {
        const createMsg: MsgCreateVideoUpscalerTaskEncodeObject = {
            typeUrl: typeUrlMsgCreateVideoUpscalerTask,
            value: {
                creator: creator,
                 cid: cid,
                startFrame: startFrame,
                endFrame: endFrame,
                threads: threads,
                reward: reward,
                
            },
        }
        return this.signAndBroadcast(creator, [createMsg],fee)
    }

    
}