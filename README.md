# comfyui-go
这是一个使用go编写的comfyui的client，旨在提交任务到comfyui，并自动完成从comfyui提取生成的图片。

众所周知，comfyui最近很火，它可以实现许多Stable Diffusion WebUI所无法实现的功能，并且生成的图也更具细节。所以本人写个破轮子拿来给各位玩玩看看，代码并不太严谨，只是能跑的起来的地步，欢迎提issue来表明你的意见和建议，如果发现bug也可以直接提issue，我会修复它。

This is a go-written client for comfyui, which intends to submit tasks to comfyui and automatically complete the extraction of generated images from comfyui.

As we all know, comfyui is very hot recently, it can implement many functions that Stable Diffusion WebUI cannot,
and the generated images are more detailed. So I wrote a piece of junk to play around with, the code is not very rigorous, but it can run to the bottom. Welcome to submit issues to indicate your opinions, if you found any bugs, you can directly submit an issue, I will fix it.




# Examples

## First
implements this interface `comfy_tasks/AigcTask` , you can find the example in `examples` directory, this interface as the task handler in the task process.

```go
package main


func main() {

	// put the comfyui api.json here 
	comfyParams := "{\n  \"3\": {\n    \"inputs\": {\n      \"seed\": 1092496014435487,\n      \"steps\": 20,\n      \"cfg\": 8,\n      \"sampler_name\": \"euler\",\n      \"scheduler\": \"normal\",\n      \"denoise\": 1,\n      \"model\": [\n        \"4\",\n        0\n      ],\n      \"positive\": [\n        \"6\",\n        0\n      ],\n      \"negative\": [\n        \"7\",\n        0\n      ],\n      \"latent_image\": [\n        \"5\",\n        0\n      ]\n    },\n    \"class_type\": \"KSampler\",\n    \"_meta\": {\n      \"title\": \"K采样器\"\n    }\n  },\n  \"4\": {\n    \"inputs\": {\n      \"ckpt_name\": \"dreamshaperXL_sfwLightningDPMSDE.safetensors\"\n    },\n    \"class_type\": \"CheckpointLoaderSimple\",\n    \"_meta\": {\n      \"title\": \"Checkpoint加载器(简易)\"\n    }\n  },\n  \"5\": {\n    \"inputs\": {\n      \"width\": 512,\n      \"height\": 512,\n      \"batch_size\": 1\n    },\n    \"class_type\": \"EmptyLatentImage\",\n    \"_meta\": {\n      \"title\": \"空Latent\"\n    }\n  },\n  \"6\": {\n    \"inputs\": {\n      \"text\": \"a girl\",\n      \"clip\": [\n        \"4\",\n        1\n      ]\n    },\n    \"class_type\": \"CLIPTextEncode\",\n    \"_meta\": {\n      \"title\": \"CLIP文本编码器\"\n    }\n  },\n  \"7\": {\n    \"inputs\": {\n      \"text\": \"text, watermark\",\n      \"clip\": [\n        \"4\",\n        1\n      ]\n    },\n    \"class_type\": \"CLIPTextEncode\",\n    \"_meta\": {\n      \"title\": \"CLIP文本编码器\"\n    }\n  },\n  \"8\": {\n    \"inputs\": {\n      \"samples\": [\n        \"3\",\n        0\n      ],\n      \"vae\": [\n        \"4\",\n        2\n      ]\n    },\n    \"class_type\": \"VAEDecode\",\n    \"_meta\": {\n      \"title\": \"VAE解码\"\n    }\n  },\n  \"10\": {\n    \"inputs\": {\n      \"images\": [\n        \"8\",\n        0\n      ]\n    },\n    \"class_type\": \"PreviewImage\",\n    \"_meta\": {\n      \"title\": \"预览图像\"\n    }\n  }\n}"
	taskHandler := New("1", comfyParams)
	handler := comfy_tasks.NewAigcTaskProcessor(context.Background(),
		"http://127.0.0.1:8443",
		"ws://127.0.0.1:8443",
		taskHandler, &demoLogger{})
	err := handler.Start()
	if err != nil {
		panic(err)
	}
}

type demoLogger struct{}

func (t *demoLogger) Info(ctx context.Context, data ...interface{}) {
	fmt.Println(data)
}

```
