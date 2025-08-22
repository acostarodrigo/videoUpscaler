import { expect } from "chai"
import { config } from "dotenv"
import _ from "../../environment"
import { VideoUpscalerStargateClient } from "../../src/stargateClient"
import { VideoUpscalerExtension } from "../../src/modules/queries"

config()

describe("VideoUpscaler", function () {
    let client: VideoUpscalerStargateClient, videoUpscaler: VideoUpscalerExtension["videoUpscaler"]

    before("create client", async function () {
        client = await VideoUpscalerStargateClient.connect(process.env.RPC_URL)
        videoUpscaler = client.videoUpscalerQueryClient!.videoUpscaler
    })

    it("can get videoUpscalertask", async function () {
        const task = await videoUpscaler.GetVideoUpscalerTask('1')
        expect(task.cid).to.be.equal("QmYC32RNLAMPRa8RGWEEHJWMcrnMzJ2Hq8xByupeFPUNtn")
    })


})