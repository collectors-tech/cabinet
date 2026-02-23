import type { PropsWithChildren } from "react";

type ScreenContainerProps = PropsWithChildren<{
  testId: string;
  title: string;
  id?: string;
}>;

function ScreenContainer({ testId, title, id, children }: ScreenContainerProps) {
  return (
    <section id={id} data-testid={testId} aria-label={title}>
      {children}
    </section>
  );
}

export function DashboardScreen(props: PropsWithChildren<{ id?: string }>) {
  return <ScreenContainer testId="screen-dashboard" title="Dashboard" id={props.id}>{props.children}</ScreenContainer>;
}

export function CollectionScreen(props: PropsWithChildren<{ id?: string }>) {
  return <ScreenContainer testId="screen-collection" title="Collection" id={props.id}>{props.children}</ScreenContainer>;
}

export function ScannerScreen(props: PropsWithChildren<{ id?: string }>) {
  return <ScreenContainer testId="screen-scanner" title="Scanner" id={props.id}>{props.children}</ScreenContainer>;
}

export function PricingScreen(props: PropsWithChildren<{ id?: string }>) {
  return <ScreenContainer testId="screen-pricing" title="Pricing" id={props.id}>{props.children}</ScreenContainer>;
}

export function DiscoveriesScreen(props: PropsWithChildren<{ id?: string }>) {
  return <ScreenContainer testId="screen-discoveries" title="Discoveries" id={props.id}>{props.children}</ScreenContainer>;
}

export function AIScreen(props: PropsWithChildren<{ id?: string }>) {
  return <ScreenContainer testId="screen-ai" title="AI Assist" id={props.id}>{props.children}</ScreenContainer>;
}

export function BarcodesScreen(props: PropsWithChildren<{ id?: string }>) {
  return <ScreenContainer testId="screen-barcodes" title="Barcodes" id={props.id}>{props.children}</ScreenContainer>;
}

export function PhotosScreen(props: PropsWithChildren<{ id?: string }>) {
  return <ScreenContainer testId="screen-photos" title="Photos" id={props.id}>{props.children}</ScreenContainer>;
}

export function SettingsScreen(props: PropsWithChildren<{ id?: string }>) {
  return <ScreenContainer testId="screen-settings" title="Settings" id={props.id}>{props.children}</ScreenContainer>;
}
