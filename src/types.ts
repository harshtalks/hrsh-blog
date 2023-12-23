import type { Props } from "astro";

export type Site = {
    siteUrl: string;
    author: string;
    desc: string;
    title: string;
    ogImage: string;
    keywords: string;
    postPerPage: number;
};

export type SocialMediaObjects = {
    name: SocialMediaTypes;
    href: string;
    Icon?: (_props: Props) => t;
    active: boolean;
    title: string;
}[];

export type SocialIcons = {
    [social in SocialMediaTypes]: string;
};

export type SocialMediaTypes =
    | "Github"
    | "Facebook"
    | "Instagram"
    | "LinkedIn"
    | "Mail"
    | "Twitter"
    | "YouTube"
    | "WhatsApp"
    | "Snapchat"
    | "CodePen"
    | "Discord"
    | "Cal.com";
