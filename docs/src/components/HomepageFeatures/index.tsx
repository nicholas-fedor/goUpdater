import type {ReactNode} from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  Svg: React.ComponentType<React.ComponentProps<'svg'>>;
  description: ReactNode;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'One-Command Update',
    Svg: require('@site/static/img/home/37.svg').default,
    description: (
      <>
        Perform a complete Go update cycle with a single command, handling download, installation, and verification automatically.
      </>
    ),
  },
  {
    title: 'Zero-Setup Installation',
    Svg: require('@site/static/img/home/47.svg').default,
    description: (
      <>
        Download the goUpdater binary and run it immediately without any setup or configuration required.
      </>
    ),
  },
  {
    title: 'Secure Automation',
    Svg: require('@site/static/img/home/62.svg').default,
    description: (
      <>
        Ensures security through checksum verification, automatically detects your platform, and handles privilege escalation when needed.
      </>
    ),
  },
  {
    title: 'Flexible Management',
    Svg: require('@site/static/img/home/67.svg').default,
    description: (
      <>
        Provides individual commands for granular control over download, install, uninstall, and verify operations.
      </>
    ),
  },
];

function Feature({title, Svg, description}: FeatureItem) {
  return (
    <div className={clsx('col col--3')}>
      <div className="text--center">
        <Svg className={styles.featureSvg} role="img" />
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): ReactNode {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
