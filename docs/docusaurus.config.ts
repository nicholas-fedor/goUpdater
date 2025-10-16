import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

const config: Config = {
  title: 'goUpdater',
  tagline: 'Updating Go made easy',
  favicon: 'img/favicon.ico',

  // Future flags, see https://docusaurus.io/docs/api/docusaurus-config#future
  future: {
    v4: true, // Improve compatibility with the upcoming Docusaurus v4
  },

  // Set the production url of your site here
  url: 'https://goupdater.nickfedor.com',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'nicholas-fedor', // Usually your GitHub org/user name.
  projectName: 'goUpdater', // Usually your repo name.

  onBrokenLinks: 'throw',

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          editUrl: 'https://github.com/nicholas-fedor/goUpdater/edit/main/docs/',
        },
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  plugins: [],

  themeConfig: {
    image: 'img/goUpdater.png',
    colorMode: {
      defaultMode: 'dark',
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'goUpdater',
      logo: {
        alt: 'Updating Go made easy',
        src: 'img/goUpdater.svg',
      },
      items: [
        {
          type: 'doc',
          position: 'left',
          label: 'Getting Started',
          to: '/docs/intro',
          docId: 'intro',
        },
        {
          type: 'doc',
          position: 'left',
          label: 'Commands',
          to: '/docs/commands',
          docId: 'commands',
        },
        {
          type: 'doc',
          position: 'left',
          label: 'Examples',
          to: '/docs/examples',
          docId: 'examples',
        },
        {
          type: 'doc',
          position: 'left',
          label: 'Troubleshooting',
          to: '/docs/troubleshooting',
          docId: 'troubleshooting',
        },
        {
          type: 'doc',
          position: 'left',
          label: 'Configuration',
          to: '/docs/configuration',
          docId: 'configuration',
        },
        {
          href: 'https://github.com/nicholas-fedor/goupdater',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Docs',
          items: [
            {
              label: 'Documentation',
              to: '/docs/intro',
            },
          ],
        },
        {
          title: 'Project',
          items: [
            {
              label: 'GitHub Repository',
              href: 'https://github.com/nicholas-fedor/goUpdater',
            },
            {
              label: 'Releases',
              href: 'https://github.com/nicholas-fedor/goUpdater/releases',
            },
            {
              label: 'Issues',
              href: 'https://github.com/nicholas-fedor/goUpdater/issues',
            },
            {
              label: 'Changelog',
              href: 'https://github.com/nicholas-fedor/goUpdater/blob/main/CHANGELOG.md',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Nicholas Fedor`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
