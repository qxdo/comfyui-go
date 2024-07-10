package main

import (
	"context"
	"fmt"
	"github.com/qxdo/comfyui-go/comfy_tasks"
)

func main() {

	comfyParams := "{\n  \"3\": {\n    \"inputs\": {\n      \"seed\": 1092496014435487,\n      \"steps\": 20,\n      \"cfg\": 8,\n      \"sampler_name\": \"euler\",\n      \"scheduler\": \"normal\",\n      \"denoise\": 1,\n      \"model\": [\n        \"4\",\n        0\n      ],\n      \"positive\": [\n        \"6\",\n        0\n      ],\n      \"negative\": [\n        \"7\",\n        0\n      ],\n      \"latent_image\": [\n        \"5\",\n        0\n      ]\n    },\n    \"class_type\": \"KSampler\",\n    \"_meta\": {\n      \"title\": \"K采样器\"\n    }\n  },\n  \"4\": {\n    \"inputs\": {\n      \"ckpt_name\": \"dreamshaperXL_sfwLightningDPMSDE.safetensors\"\n    },\n    \"class_type\": \"CheckpointLoaderSimple\",\n    \"_meta\": {\n      \"title\": \"Checkpoint加载器(简易)\"\n    }\n  },\n  \"5\": {\n    \"inputs\": {\n      \"width\": 512,\n      \"height\": 512,\n      \"batch_size\": 1\n    },\n    \"class_type\": \"EmptyLatentImage\",\n    \"_meta\": {\n      \"title\": \"空Latent\"\n    }\n  },\n  \"6\": {\n    \"inputs\": {\n      \"text\": \"a girl\",\n      \"clip\": [\n        \"4\",\n        1\n      ]\n    },\n    \"class_type\": \"CLIPTextEncode\",\n    \"_meta\": {\n      \"title\": \"CLIP文本编码器\"\n    }\n  },\n  \"7\": {\n    \"inputs\": {\n      \"text\": \"text, watermark\",\n      \"clip\": [\n        \"4\",\n        1\n      ]\n    },\n    \"class_type\": \"CLIPTextEncode\",\n    \"_meta\": {\n      \"title\": \"CLIP文本编码器\"\n    }\n  },\n  \"8\": {\n    \"inputs\": {\n      \"samples\": [\n        \"3\",\n        0\n      ],\n      \"vae\": [\n        \"4\",\n        2\n      ]\n    },\n    \"class_type\": \"VAEDecode\",\n    \"_meta\": {\n      \"title\": \"VAE解码\"\n    }\n  },\n  \"10\": {\n    \"inputs\": {\n      \"images\": [\n        \"8\",\n        0\n      ]\n    },\n    \"class_type\": \"PreviewImage\",\n    \"_meta\": {\n      \"title\": \"预览图像\"\n    }\n  }\n}"
	taskHandler := New("1", comfyParams)
	handler := comfy_tasks.NewAigcTaskProcessor(context.Background(),
		"http://127.0.0.1:8443",
		"ws://127.0.0.1:8443",
		taskHandler, &demoLog{})
	err := handler.Start()
	if err != nil {
		panic(err)
	}
}

type demoLog struct{}

func (t *demoLog) Info(ctx context.Context, data ...interface{}) {
	fmt.Println(data)
}
