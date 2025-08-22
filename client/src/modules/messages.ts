import { EncodeObject, GeneratedType } from "@cosmjs/proto-signing"
import {
    MsgCreateVideoUpscalerTask,
    MsgCreateVideoUpscalerTaskResponse
} from "../types/generated/janction/videoUpscaler/v1/tx"

export const typeUrlMsgCreateVideoUpscalerTask = "/janction.videoUpscaler.v1.MsgCreateVideoUpscalerTask"
export const typeUrlMsgCreateVideoUpscalerTaskResponse = "/janction.videoUpscaler.v1.MsgCreateVideoUpscalerTaskResponse"

export const videoUpscalerTypes: ReadonlyArray<[string, GeneratedType]> = [
    [typeUrlMsgCreateVideoUpscalerTask, MsgCreateVideoUpscalerTask],
    [typeUrlMsgCreateVideoUpscalerTaskResponse, MsgCreateVideoUpscalerTaskResponse],
    
]

export interface MsgCreateVideoUpscalerTaskEncodeObject extends EncodeObject {
    readonly typeUrl: "/janction.videoUpscaler.v1.MsgCreateVideoUpscalerTask"
    readonly value: Partial<MsgCreateVideoUpscalerTask>
}

export function isMsgCreateVideoUpscalerTaskEncodeObject(
    encodeObject: EncodeObject,
): encodeObject is MsgCreateVideoUpscalerTaskEncodeObject {
    return (encodeObject as MsgCreateVideoUpscalerTaskEncodeObject).typeUrl === typeUrlMsgCreateVideoUpscalerTask
}

export interface MsgCreateVideoUpscalerTaskResponseEncodeObject extends EncodeObject {
    readonly typeUrl: "/janction.videoUpscaler.v1.MsgCreateVideoUpscalerTaskResponse"
    readonly value: Partial<MsgCreateVideoUpscalerTaskResponse>
}

export function isMsgCreateVideoUpscalerTaskResponseEncodeObject(
    encodeObject: EncodeObject,
): encodeObject is MsgCreateVideoUpscalerTaskResponseEncodeObject {
    return (encodeObject as MsgCreateVideoUpscalerTaskResponseEncodeObject).typeUrl === typeUrlMsgCreateVideoUpscalerTaskResponse
}

