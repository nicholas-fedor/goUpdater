import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import HomepageFeatures from '@site/src/components/HomepageFeatures';
import Heading from '@theme/Heading';

import styles from './index.module.css';

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <Heading as="h1" className="hero__title">
          {siteConfig.title}
        </Heading>
        <p className="hero__subtitle text--italic text--bold">Updating<img src="/img/home/Go-Logo_Black.svg" style={{ height: '0.75em', width: 'auto', verticalAlign: 'baseline' }} alt=""/>made easy</p>
        <div className={styles.installation}>
          <Link
            className="button button--outline button--lg shadow--md"
            to="https://github.com/nicholas-fedor/goUpdater/releases/latest">
            <div style={{display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
              <img src="/img/home/48.svg" alt="" style={{ height: '5em', width: 'auto', verticalAlign: 'baseline' }}/>
              <p>Get the latest release</p>
            </div>
          </Link>
        </div>
      </div>
    </header>
  );
}

export default function Home(): ReactNode {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={siteConfig.title}
      description="A secure and automated Go version manager for Linux, macOS, and Windows. Install, update, and manage multiple Go versions with ease.">
      <HomepageHeader />
      <main>
        <HomepageFeatures />
      </main>
    </Layout>
  );
}
