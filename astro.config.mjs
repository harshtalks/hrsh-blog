import { defineConfig } from "astro/config";
import mdx from "@astrojs/mdx";
import sitemap from "@astrojs/sitemap";
import tailwind from "@astrojs/tailwind";
import react from "@astrojs/react";
import { SITE } from "./src/config.ts";
// https://astro.build/config
export default defineConfig({
    site: SITE.siteUrl,
    markdown: {
        shikiConfig: { theme: "min-dark" }
    },
    vite: {
        optimizeDeps: {
            exclude: ["@resvg/resvg-js"]
        },
        ssr: {
            external: ["svgo"]
        }
    },
    integrations: [mdx(), sitemap(), tailwind(), react()]
});
