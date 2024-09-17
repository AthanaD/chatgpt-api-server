import { viteBundler } from "@vuepress/bundler-vite";
import { defaultTheme } from "@vuepress/theme-default";
import { defineUserConfig } from "vuepress";

export default defineUserConfig({
  bundler: viteBundler(),
  theme: defaultTheme({
    sidebar: {
      '/guide/': [
        {
          text: '指南',
          children: [
            '/guide/README.md',
            '/guide/modelmap.md',
          ],
        },
      ]
    },
  }),

  lang: "zh-CN",
  title: "ChatgptApiServer",
  description: "快捷的chat2api服务端",
});
