import { useEffect, useRef, useState } from "react";
import { RecoveryPassphraseForm, SessionTokenForm } from "./components/auth-forms";
import { CollectionItemForm, type CollectionItemValues } from "./components/collection-item-form";
import { DataImportExportWizard } from "./components/data-import-export-wizard";
import { InstanceForm, type InstanceFormValues } from "./components/instance-form";
import { ScannerQuerySetForm, type ScannerQuerySetValues } from "./components/scanner-query-set-form";
import { AIAssistForms } from "./components/ai-assist-forms";
import { ProfileSettingsForm, SecretsForm, type ProfileSettingsValues, type SecretsValues } from "./components/settings-secrets-forms";
import { StarterQuickAddForm, type StarterQuickAddValues } from "./components/starter-quick-add-form";
import { CollectionScreen, DashboardScreen, PricingScreen, ScannerScreen, SettingsScreen } from "./screens/top-level-screens";

type Theme = "light" | "dark";
type OnboardingStep = 1 | 2 | 3 | 4 | 5;
type TopLevelScreen = "all" | "dashboard" | "collection" | "scanner" | "pricing" | "settings";
type ScannerQuerySetRecord = {
  id: string;
  name: string;
  keywords?: string[];
  exclusions?: string[];
  max_price?: number;
  region?: string;
  condition?: string;
  schedule_cron?: string;
  enabled?: boolean;
  rate_limit_rps?: number;
  max_retry_count?: number;
};

function detectInitialTheme(): Theme {
  const saved = localStorage.getItem("cabinet.theme");
  if (saved === "dark" || saved === "light") {
    return saved;
  }
  return "light";
}

export function App() {
  const ONBOARDING_STEPS = ["Welcome", "Identity", "Starter Data", "First Item", "Preferences"] as const;
  const [theme, setTheme] = useState<Theme>(detectInitialTheme);
  const [profiles, setProfiles] = useState<Array<{ id: string; name: string }>>([]);
  const [activeProfile, setActiveProfile] = useState<{ id: string; name: string } | null>(null);
  const [profileStorage, setProfileStorage] = useState<{ db_path?: string; media_dir?: string } | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [items, setItems] = useState<Array<{ id: string; part_number: string; title: string; brand?: string; category?: string; series?: string }>>(
    [],
  );
  const [itemsLoading, setItemsLoading] = useState(false);
  const [starterSubmitting, setStarterSubmitting] = useState(false);
  const [itemsError, setItemsError] = useState("");
  const [collectionFormSeed, setCollectionFormSeed] = useState<Partial<CollectionItemValues>>({
    part_number: "",
    title: "",
    brand: "",
    category: "General",
  });
  const [collectionFormKey, setCollectionFormKey] = useState(0);
  const [instances, setInstances] = useState<
    Array<{ id: string; item_id: string; condition?: string; status?: string; quantity?: number; storage_location?: string; acquisition_price?: number; acquisition_date?: string }>
  >([]);
  const [instanceSubmitting, setInstanceSubmitting] = useState(false);
  const [collectionQuery, setCollectionQuery] = useState({
    text: "",
    brand: "",
    category: "",
    sort_by: "date_added",
  });
  const [savedFilters, setSavedFilters] = useState<Array<{ id: string; name: string; query?: Record<string, unknown> }>>([]);
  const [savedFilterName, setSavedFilterName] = useState("");
  const [columnBrand, setColumnBrand] = useState("");
  const [columnCategory, setColumnCategory] = useState("");
  const [columnSeries, setColumnSeries] = useState("");
  const [selectedItemID, setSelectedItemID] = useState("");
  const [photos, setPhotos] = useState<Array<{ id: string; item_id: string; filename: string; is_primary?: boolean }>>([]);
  const [photoFile, setPhotoFile] = useState<File | null>(null);
  const [fullscreenPhoto, setFullscreenPhoto] = useState<{ id: string; filename: string } | null>(null);
  const [cameraOpen, setCameraOpen] = useState(false);
  const [cameraStatus, setCameraStatus] = useState("idle");
  const [photosError, setPhotosError] = useState("");
  const [querySets, setQuerySets] = useState<ScannerQuerySetRecord[]>([]);
  const [selectedQuerySetID, setSelectedQuerySetID] = useState("");
  const [editingQuerySetID, setEditingQuerySetID] = useState("");
  const [querySetSubmitting, setQuerySetSubmitting] = useState(false);
  const [scheduledRunStatus, setScheduledRunStatus] = useState("");
  const [scannerRetryStatus, setScannerRetryStatus] = useState("");
  const [scannerFailures, setScannerFailures] = useState<Array<{ id?: string; query_set_id?: string; reason?: string; attempts?: number; last_error_at?: string }>>([]);
  const [providerHealth, setProviderHealth] = useState<{ provider?: string; state?: string; healthy?: boolean } | null>(null);
  const [matchingRunStatus, setMatchingRunStatus] = useState("");
  const [candidates, setCandidates] = useState<Array<{ id: string; title?: string; listing_id?: string; status?: string }>>([]);
  const [matchingResults, setMatchingResults] = useState<Array<{ candidate_id?: string; state?: string; part_number?: string; item_id?: string }>>(
    [],
  );
  const [notInCollectionItems, setNotInCollectionItems] = useState<Array<{ candidate_id: string; title?: string; price?: number; url?: string; last_seen?: string }>>(
    [],
  );
  const [notInCollectionFilter, setNotInCollectionFilter] = useState({ query: "", maxPrice: "", dateFrom: "" });
  const [scannerError, setScannerError] = useState("");
  const [dashboard, setDashboard] = useState<Record<string, unknown> | null>(null);
  const [dashboardLoading, setDashboardLoading] = useState(false);
  const [wishlist, setWishlist] = useState<Array<{ id: string; item_id: string; target_price?: number; below_target_now?: boolean; priority?: string }>>([]);
  const [pricingPoints, setPricingPoints] = useState<Array<{ day?: string; date?: string; price?: number; min?: number; median?: number; latest?: number }>>([]);
  const [pricingBySource, setPricingBySource] = useState<Record<string, Array<{ snapshot_date?: string; min_price?: number; median_price?: number; latest_price?: number }>>>(
    {},
  );
  const [pricingHistory, setPricingHistory] = useState<Array<{ day?: string; date?: string; min?: number; median?: number; latest?: number }>>([]);
  const [pricingStats, setPricingStats] = useState<Record<string, unknown> | null>(null);
  const [pricingTrend, setPricingTrend] = useState<Record<string, unknown> | null>(null);
  const [wishlistHits, setWishlistHits] = useState<Array<{ item_id?: string; listing_id?: string; title?: string; price?: number }>>([]);
  const [pricingTrackStatus, setPricingTrackStatus] = useState("");
  const [snapshotStatus, setSnapshotStatus] = useState("");
  const [wishlistDraft, setWishlistDraft] = useState({ item_id: "", target_price: "0", priority: "normal", notes: "" });
  const [insightError, setInsightError] = useState("");
  const [licenseStatus, setLicenseStatus] = useState<{ state?: string; tier?: string; features?: string[]; expires_at?: string } | null>(null);
  const [profileLicenseJSON, setProfileLicenseJSON] = useState("");
  const [licenseImportDraft, setLicenseImportDraft] = useState({ payload_base64: "", signature_base64: "" });
  const [licenseImportStatus, setLicenseImportStatus] = useState("");
  const [debugModeEnabled, setDebugModeEnabled] = useState(false);
  const [logCount, setLogCount] = useState(0);
  const [activityLogs, setActivityLogs] = useState<Array<Record<string, unknown>>>([]);
  const [runtimeDiagnostics, setRuntimeDiagnostics] = useState<{ update_channel?: string; update_public_key_configured?: boolean } | null>(null);
  const [recoveryDiagnostics, setRecoveryDiagnostics] = useState<{ recovery_required?: boolean } | null>(null);
  const [backupEntries, setBackupEntries] = useState<Array<{ path: string; name: string; timestampLabel: string }>>([]);
  const [selectedBackupPath, setSelectedBackupPath] = useState("");
  const [confirmRestore, setConfirmRestore] = useState(false);
  const [adminError, setAdminError] = useState("");
  const [settingsStatus, setSettingsStatus] = useState("");
  const [profileSettingsInitial, setProfileSettingsInitial] = useState<ProfileSettingsValues>({
    scanner_schedule: "",
    backup_frequency: "daily",
    db_path: "",
    update_channel: "stable",
  });
  const [secretsInitial, setSecretsInitial] = useState<Partial<SecretsValues>>({
    openai_api_key: "",
    ebay_app_id: "",
    ebay_auth_token: "",
  });
  const [exportBytes, setExportBytes] = useState(0);
  const [barcodes, setBarcodes] = useState<Array<{ id?: string; barcode: string }>>([]);
  const [barcodeInput, setBarcodeInput] = useState("");
  const [barcodeLookupMatches, setBarcodeLookupMatches] = useState<Array<{ item_id?: string; part_number?: string }>>([]);
  const [barcodeExternalURL, setBarcodeExternalURL] = useState("");
  const [barcodeError, setBarcodeError] = useState("");
  const [aiEnabled, setAiEnabled] = useState(false);
  const [aiTitleInput, setAiTitleInput] = useState("");
  const [aiPhotoURL, setAiPhotoURL] = useState("");
  const [aiSuggestion, setAiSuggestion] = useState<{ title?: string; confidence?: number; [key: string]: unknown } | null>(null);
  const [aiError, setAiError] = useState("");
  const [aiLastAction, setAiLastAction] = useState<"title" | "photo" | "">("");
  const [authStatus, setAuthStatus] = useState("");
  const [starterIdentityBusy, setStarterIdentityBusy] = useState(false);
  const [recoverySubmitting, setRecoverySubmitting] = useState(false);
  const [authSessionID, setAuthSessionID] = useState("");
  const [requiresRegistration, setRequiresRegistration] = useState<boolean | null>(null);
  const [onboardingStatus, setOnboardingStatus] = useState("");
  const [onboardingStep, setOnboardingStep] = useState<OnboardingStep>(1);
  const [onboardingCompleted, setOnboardingCompleted] = useState(false);
  const [onboardingIdentityComplete, setOnboardingIdentityComplete] = useState(false);
  const [onboardingSetupPath, setOnboardingSetupPath] = useState<"quick" | "import" | "sample" | "">("");
  const [onboardingStarterDataChoice, setOnboardingStarterDataChoice] = useState<"sample" | "empty" | "">("");
  const [onboardingBackupFrequency, setOnboardingBackupFrequency] = useState<"daily" | "weekly" | "monthly">("daily");
  const [onboardingScannerPreset, setOnboardingScannerPreset] = useState<"manual" | "daily" | "weekly">("manual");
  const [onboardingTheme, setOnboardingTheme] = useState<Theme>(detectInitialTheme);
  const [onboardingFinishing, setOnboardingFinishing] = useState(false);
  const [advancedWorkspace, setAdvancedWorkspace] = useState(false);
  const [activeScreen, setActiveScreen] = useState<TopLevelScreen>("all");
  const [credentialJSON, setCredentialJSON] = useState("{}");
  const [sessionToken, setSessionToken] = useState("");
  const [recoveryPassphrase, setRecoveryPassphrase] = useState("");
  const [mobileNavOpen, setMobileNavOpen] = useState(false);
  const drawerRef = useRef<HTMLDivElement | null>(null);
  const drawerTriggerRef = useRef<HTMLButtonElement | null>(null);

  useEffect(() => {
    document.body.setAttribute("data-theme", theme);
    localStorage.setItem("cabinet.theme", theme);
  }, [theme]);

  useEffect(() => {
    if (!mobileNavOpen) {
      return;
    }
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    drawerRef.current?.focus();

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setMobileNavOpen(false);
        return;
      }
      if (event.key !== "Tab") {
        return;
      }
      const root = drawerRef.current;
      if (!root) {
        return;
      }
      const focusable = Array.from(
        root.querySelectorAll<HTMLElement>('a[href], button:not([disabled]), [tabindex]:not([tabindex="-1"])'),
      );
      if (focusable.length === 0) {
        event.preventDefault();
        return;
      }
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      const active = document.activeElement as HTMLElement | null;
      if (event.shiftKey && active === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && active === last) {
        event.preventDefault();
        first.focus();
      }
    };
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.body.style.overflow = previousOverflow;
      document.removeEventListener("keydown", onKeyDown);
      drawerTriggerRef.current?.focus();
    };
  }, [mobileNavOpen]);

  useEffect(() => {
    if (!activeProfile?.id || !advancedWorkspace || activeScreen !== "dashboard") {
      return;
    }
    void loadDashboard();
  }, [activeProfile?.id, advancedWorkspace, activeScreen]);

  useEffect(() => {
    let disposed = false;
    async function loadProfiles() {
      setLoading(true);
      setError("");
      try {
        const resp = await fetch("/api/profiles");
        if (!resp.ok) {
          throw new Error("failed_to_list_profiles");
        }
        const data = (await resp.json()) as { profiles?: Array<{ id: string; name: string }> };
        if (disposed) {
          return;
        }
        setProfiles(data.profiles ?? []);
      } catch (e) {
        if (disposed) {
          return;
        }
        setError(e instanceof Error ? e.message : "failed_to_load_profiles");
      } finally {
        if (!disposed) {
          setLoading(false);
        }
      }
    }
    loadProfiles();
    return () => {
      disposed = true;
    };
  }, []);

  async function loadProfileStorage(profileID: string) {
    const storageResp = await fetch(`/api/profiles/${profileID}/storage`);
    if (!storageResp.ok) {
      throw new Error("failed_to_load_profile_storage");
    }
    const storage = (await storageResp.json()) as { db_path?: string; media_dir?: string };
    setProfileStorage(storage);
  }

  async function loadItems() {
    setItemsLoading(true);
    setItemsError("");
    try {
      const resp = await fetch("/api/items");
      if (!resp.ok) {
        throw new Error("failed_to_list_items");
      }
      const data = (await resp.json()) as { items?: Array<{ id: string; part_number: string; title: string; brand?: string; category?: string; series?: string }> };
      const listed = data.items || [];
      setItems(listed);
      if (listed.length > 0 && !selectedItemID) {
        setSelectedItemID(listed[0].id);
      }
    } catch (e) {
      setItemsError(e instanceof Error ? e.message : "failed_to_list_items");
    } finally {
      setItemsLoading(false);
    }
  }

  async function searchItems() {
    setItemsLoading(true);
    setItemsError("");
    try {
      const params = new URLSearchParams();
      if (collectionQuery.text.trim()) {
        params.set("q", collectionQuery.text.trim());
      }
      if (collectionQuery.brand.trim()) {
        params.set("brand", collectionQuery.brand.trim());
      }
      if (collectionQuery.category.trim()) {
        params.set("category", collectionQuery.category.trim());
      }
      if (collectionQuery.sort_by.trim()) {
        params.set("sort", collectionQuery.sort_by.trim());
      }
      const resp = await fetch(`/api/search/items?${params.toString()}`);
      if (!resp.ok) {
        throw new Error("failed_to_search_items");
      }
      const data = (await resp.json()) as { items?: Array<{ id: string; part_number: string; title: string; brand?: string; category?: string; series?: string }> };
      const listed = data.items || [];
      setItems(listed);
      if (listed.length > 0 && !selectedItemID) {
        setSelectedItemID(listed[0].id);
      }
    } catch (e) {
      setItemsError(e instanceof Error ? e.message : "failed_to_search_items");
    } finally {
      setItemsLoading(false);
    }
  }

  useEffect(() => {
    if (!activeProfile?.id) {
      return;
    }
    const hasSearch =
      collectionQuery.text.trim() ||
      collectionQuery.brand.trim() ||
      collectionQuery.category.trim() ||
      collectionQuery.sort_by !== "date_added";
    if (!hasSearch) {
      return;
    }
    const handle = window.setTimeout(() => {
      void searchItems();
    }, 300);
    return () => window.clearTimeout(handle);
  }, [activeProfile?.id, collectionQuery.text, collectionQuery.brand, collectionQuery.category, collectionQuery.sort_by]);

  useEffect(() => {
    let disposed = false;
    async function loadRequirements() {
      if (!activeProfile?.id) {
        setRequiresRegistration(null);
        return;
      }
      try {
        const reqResp = await fetch(`/api/auth/requirements?profile_id=${encodeURIComponent(activeProfile.id)}`);
        if (!reqResp.ok) {
          throw new Error("failed_to_get_auth_requirements");
        }
        const reqs = (await reqResp.json()) as { requires_registration?: boolean };
        if (!disposed) {
          setRequiresRegistration(Boolean(reqs.requires_registration));
        }
      } catch {
        if (!disposed) {
          setRequiresRegistration(null);
        }
      }
    }
    loadRequirements();
    return () => {
      disposed = true;
    };
  }, [activeProfile?.id]);

  useEffect(() => {
    if (!activeProfile?.id || requiresRegistration !== true) {
      return;
    }
    const key = workspacePreferenceKey(activeProfile.id);
    const onboardingCompletedValue = localStorage.getItem(onboardingCompletedPreferenceKey(activeProfile.id));
    const onboardingStepValue = localStorage.getItem(onboardingStepPreferenceKey(activeProfile.id));
    const hasOnboardingState = onboardingCompletedValue !== null || onboardingStepValue !== null;
    const onboardingComplete = onboardingCompletedValue === "1" || onboardingCompletedValue?.toLowerCase() === "true";
    if (onboardingComplete) {
      return;
    }
    setAdvancedWorkspace(false);
    localStorage.setItem(key, "0");
    if (!hasOnboardingState) {
      localStorage.setItem(onboardingCompletedPreferenceKey(activeProfile.id), "0");
      localStorage.setItem(onboardingStepPreferenceKey(activeProfile.id), "1");
    }
  }, [activeProfile?.id, requiresRegistration]);

  async function postJSON(path: string, payload: unknown) {
    const resp = await fetch(path, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!resp.ok) {
      throw new Error(`request_failed:${path}`);
    }
    return (await resp.json()) as Record<string, unknown>;
  }

  async function createFirstProfile() {
    setError("");
    try {
      const createResp = await fetch("/api/profiles", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: "Default" }),
      });
      if (!createResp.ok) {
        throw new Error("failed_to_create_profile");
      }
      const created = (await createResp.json()) as { id: string; name: string };
      setProfiles([created]);

      const activateResp = await fetch("/api/profiles/active", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ profile_id: created.id }),
      });
      if (!activateResp.ok) {
        throw new Error("failed_to_activate_profile");
      }
      const active = (await activateResp.json()) as { id: string; name: string };
      setActiveProfile(active);
      await loadProfileStorage(active.id);
      localStorage.setItem(workspacePreferenceKey(active.id), "0");
      setAdvancedWorkspace(false);
      setOnboardingStep(1);
      setOnboardingCompleted(false);
      setOnboardingIdentityComplete(false);
      setOnboardingSetupPath("");
      setOnboardingStarterDataChoice("");
      setOnboardingBackupFrequency("daily");
      setOnboardingScannerPreset("manual");
      setOnboardingTheme(theme);
      localStorage.setItem(onboardingStepPreferenceKey(active.id), "1");
      localStorage.setItem(onboardingCompletedPreferenceKey(active.id), "0");
      localStorage.setItem(onboardingIdentityPreferenceKey(active.id), "0");
      localStorage.removeItem(onboardingPathPreferenceKey(active.id));
      localStorage.removeItem(onboardingStarterDataPreferenceKey(active.id));
      await loadItems();
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_setup_profile");
    }
  }

  async function activateProfile(profileID: string) {
    setError("");
    try {
      const activateResp = await fetch("/api/profiles/active", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ profile_id: profileID }),
      });
      if (!activateResp.ok) {
        throw new Error("failed_to_activate_profile");
      }
      const active = (await activateResp.json()) as { id: string; name: string };
      setActiveProfile(active);
      await loadProfileStorage(active.id);
      await loadOnboardingState(active.id);
      await loadItems();
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_activate_profile");
    }
  }

  async function loadOnboardingState(profileID: string) {
    const onboardingCompletedValue = localStorage.getItem(onboardingCompletedPreferenceKey(profileID));
    const onboardingStepValue = localStorage.getItem(onboardingStepPreferenceKey(profileID));
    const onboardingPathValue = localStorage.getItem(onboardingPathPreferenceKey(profileID));
    const onboardingIdentityValue = localStorage.getItem(onboardingIdentityPreferenceKey(profileID));
    const onboardingStarterDataValue = localStorage.getItem(onboardingStarterDataPreferenceKey(profileID));
    const workspaceValue = localStorage.getItem(workspacePreferenceKey(profileID));
    const hasExplicitOnboardingState = onboardingCompletedValue !== null || onboardingStepValue !== null;
    const completed =
      onboardingCompletedValue === "1" ||
      onboardingCompletedValue?.toLowerCase() === "true" ||
      (!hasExplicitOnboardingState && workspaceValue === "1");
    setOnboardingCompleted(Boolean(completed));

    const savedStep = Number(onboardingStepValue || "1");
    if (Number.isFinite(savedStep) && savedStep >= 1 && savedStep <= ONBOARDING_STEPS.length) {
      setOnboardingStep(savedStep as OnboardingStep);
    } else {
      setOnboardingStep(1);
    }
    if (onboardingPathValue === "quick" || onboardingPathValue === "import" || onboardingPathValue === "sample") {
      setOnboardingSetupPath(onboardingPathValue);
    } else {
      setOnboardingSetupPath("");
    }
    setOnboardingIdentityComplete(onboardingIdentityValue === "1" || onboardingIdentityValue?.toLowerCase() === "true");
    if (onboardingStarterDataValue === "sample" || onboardingStarterDataValue === "empty") {
      setOnboardingStarterDataChoice(onboardingStarterDataValue);
    } else {
      setOnboardingStarterDataChoice("");
    }
    setOnboardingBackupFrequency("daily");
    setOnboardingScannerPreset("manual");
    setOnboardingTheme(theme);

    if (!completed && !hasExplicitOnboardingState && workspaceValue === null) {
      setAdvancedWorkspace(true);
      return;
    }

    if (!completed) {
      setAdvancedWorkspace(false);
      localStorage.setItem(workspacePreferenceKey(profileID), "0");
      return;
    }

    if (workspaceValue === null) {
      setAdvancedWorkspace(false);
      localStorage.setItem(workspacePreferenceKey(profileID), "0");
      return;
    }
    setAdvancedWorkspace(workspaceValue === "1" || workspaceValue.toLowerCase() === "true");
  }

  async function setWorkspaceMode(nextAdvanced: boolean) {
    if (!activeProfile?.id) {
      return;
    }
    setAdvancedWorkspace(nextAdvanced);
    if (nextAdvanced) {
      setActiveScreen("all");
    }
    localStorage.setItem(workspacePreferenceKey(activeProfile.id), nextAdvanced ? "1" : "0");
    if (nextAdvanced) {
      setOnboardingCompleted(true);
      setOnboardingStep(5);
      localStorage.setItem(onboardingStepPreferenceKey(activeProfile.id), "5");
      localStorage.setItem(onboardingCompletedPreferenceKey(activeProfile.id), "1");
    }
  }

  function workspacePreferenceKey(profileID: string) {
    return `cabinet.workspace.${profileID}`;
  }

  function onboardingStepPreferenceKey(profileID: string) {
    return `cabinet.onboarding.step.${profileID}`;
  }

  function onboardingCompletedPreferenceKey(profileID: string) {
    return `cabinet.onboarding.completed.${profileID}`;
  }

  function onboardingPathPreferenceKey(profileID: string) {
    return `cabinet.onboarding.path.${profileID}`;
  }

  function onboardingIdentityPreferenceKey(profileID: string) {
    return `cabinet.onboarding.identity_completed.${profileID}`;
  }

  function onboardingStarterDataPreferenceKey(profileID: string) {
    return `cabinet.onboarding.starter_data.${profileID}`;
  }

  async function chooseOnboardingPath(path: "quick" | "import" | "sample") {
    if (!activeProfile?.id) {
      return;
    }
    setOnboardingSetupPath(path);
    localStorage.setItem(onboardingPathPreferenceKey(activeProfile.id), path);
    if (path === "sample") {
      await seedOnboardingSampleData();
    }
    if (path === "import") {
      setOnboardingStatus("Import path selected. Continue to setup identity, then use import tools in advanced workspace.");
    }
    const next: OnboardingStep = 2;
    setOnboardingStep(next);
    localStorage.setItem(onboardingStepPreferenceKey(activeProfile.id), String(next));
  }

  async function chooseOnboardingStarterData(choice: "sample" | "empty") {
    if (!activeProfile?.id) {
      return;
    }
    setOnboardingStarterDataChoice(choice);
    localStorage.setItem(onboardingStarterDataPreferenceKey(activeProfile.id), choice);
    if (choice === "sample") {
      await seedOnboardingSampleData();
      return;
    }
    setOnboardingStatus("Starting with an empty collection.");
  }

  function scannerScheduleFromPreset(preset: "manual" | "daily" | "weekly") {
    if (preset === "daily") {
      return "0 8 * * *";
    }
    if (preset === "weekly") {
      return "0 8 * * 1";
    }
    return "";
  }

  async function finishOnboarding() {
    if (!activeProfile?.id) {
      return;
    }
    setOnboardingFinishing(true);
    setAdminError("");
    try {
      const payload = {
        settings: {
          scanner_schedule: scannerScheduleFromPreset(onboardingScannerPreset),
          backup_frequency: onboardingBackupFrequency,
          "storage.db_path": profileSettingsInitial.db_path || "",
          update_channel: profileSettingsInitial.update_channel || "stable",
        },
      };
      const resp = await fetch(`/api/profiles/${encodeURIComponent(activeProfile.id)}/settings`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!resp.ok) {
        throw new Error("failed_to_update_settings");
      }
      setTheme(onboardingTheme);
      setSettingsStatus("settings_saved");
      setOnboardingStatus("Onboarding complete. Advanced workspace unlocked.");
      await setWorkspaceMode(true);
    } catch (e) {
      setAdminError(e instanceof Error ? e.message : "failed_to_update_settings");
    } finally {
      setOnboardingFinishing(false);
    }
  }

  function nextOnboardingStep() {
    setOnboardingStep((current) => {
      const next = current < 5 ? ((current + 1) as OnboardingStep) : current;
      if (activeProfile?.id) {
        localStorage.setItem(onboardingStepPreferenceKey(activeProfile.id), String(next));
      }
      return next;
    });
  }

  function previousOnboardingStep() {
    setOnboardingStep((current) => {
      const next = current > 1 ? ((current - 1) as OnboardingStep) : current;
      if (activeProfile?.id) {
        localStorage.setItem(onboardingStepPreferenceKey(activeProfile.id), String(next));
      }
      return next;
    });
  }

  async function addItemWithPayload(payload: {
    part_number: string;
    title: string;
    brand: string;
    category: string;
    make?: string;
    model?: string;
    year?: string;
    scale?: string;
    series?: string;
    description?: string;
    tags?: string[];
  }) {
    setItemsError("");
    if (!payload.part_number.trim() || !payload.title.trim()) {
      setItemsError("part_number_and_title_required");
      return;
    }
    try {
      const resp = await fetch("/api/items", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!resp.ok) {
        throw new Error("failed_to_create_item");
      }
      const created = (await resp.json()) as { id: string; part_number: string; title: string; brand?: string; category?: string; series?: string };
      setItems((current) => [...current, created]);
      if (!selectedItemID) {
        setSelectedItemID(created.id);
      }
    } catch (e) {
      setItemsError(e instanceof Error ? e.message : "failed_to_create_item");
    }
  }

  async function addStarterItem(values: StarterQuickAddValues) {
    setStarterSubmitting(true);
    try {
      await addItemWithPayload({
        part_number: values.part_number,
        title: values.title,
        brand: values.brand || "",
        category: values.category || "General",
        series: values.series || "",
        description: values.description || "",
      });
      if (activeProfile?.id && onboardingStep === 4) {
        const next: OnboardingStep = 5;
        setOnboardingStep(next);
        localStorage.setItem(onboardingStepPreferenceKey(activeProfile.id), String(next));
        setOnboardingStatus("First item added. Continue to preferences.");
      }
    } finally {
      setStarterSubmitting(false);
    }
  }

  async function addCollectionItem(values: CollectionItemValues) {
    const tags = values.tags
      ? values.tags
          .split(",")
          .map((tag) => tag.trim())
          .filter(Boolean)
      : [];
    await addItemWithPayload({
      part_number: values.part_number,
      title: values.title,
      brand: values.brand,
      category: values.category,
      make: values.make || "",
      model: values.model || "",
      year: values.year || "",
      scale: values.scale || "",
      series: values.series || "",
      description: values.description || "",
      tags,
    });
    setCollectionFormSeed({
      part_number: "",
      title: "",
      brand: "",
      category: "General",
    });
    setCollectionFormKey((current) => current + 1);
  }

  async function loadInstances() {
    if (!selectedItemID) {
      setItemsError("item_id_required");
      return;
    }
    setItemsError("");
    try {
      const resp = await fetch(`/api/items/${encodeURIComponent(selectedItemID)}/instances`);
      if (!resp.ok) {
        throw new Error("failed_to_list_instances");
      }
      const data = (await resp.json()) as {
        instances?: Array<{
          id: string;
          item_id: string;
          condition?: string;
          status?: string;
          quantity?: number;
          storage_location?: string;
          acquisition_price?: number;
          acquisition_date?: string;
        }>;
      };
      setInstances(data.instances || []);
    } catch (e) {
      setItemsError(e instanceof Error ? e.message : "failed_to_list_instances");
    }
  }

  async function addInstance(values: InstanceFormValues) {
    if (!values.item_id.trim()) {
      setItemsError("item_id_required");
      return;
    }
    setInstanceSubmitting(true);
    setItemsError("");
    try {
      const resp = await fetch(`/api/items/${encodeURIComponent(values.item_id)}/instances`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          condition: values.condition,
          status: values.status,
          quantity: values.quantity,
          storage_location: values.storage_location || "",
          acquisition_price: values.acquisition_price || 0,
          acquisition_date: values.acquisition_date || "",
          notes: values.notes || "",
        }),
      });
      if (!resp.ok) {
        throw new Error("failed_to_create_instance");
      }
      const created = (await resp.json()) as {
        id: string;
        item_id: string;
        condition?: string;
        status?: string;
        quantity?: number;
        storage_location?: string;
        acquisition_price?: number;
        acquisition_date?: string;
      };
      setInstances((current) => [...current, created]);
    } catch (e) {
      setItemsError(e instanceof Error ? e.message : "failed_to_create_instance");
    } finally {
      setInstanceSubmitting(false);
    }
  }

  async function loadSavedFilters() {
    if (!activeProfile?.id) {
      return;
    }
    setItemsError("");
    try {
      const resp = await fetch(`/api/profiles/${encodeURIComponent(activeProfile.id)}/saved-filters`);
      if (!resp.ok) {
        throw new Error("failed_to_list_saved_filters");
      }
      const data = (await resp.json()) as { saved_filters?: Array<{ id: string; name: string; query?: Record<string, unknown> }> };
      setSavedFilters(data.saved_filters || []);
    } catch (e) {
      setItemsError(e instanceof Error ? e.message : "failed_to_list_saved_filters");
    }
  }

  async function saveCurrentFilter() {
    if (!activeProfile?.id || !savedFilterName.trim()) {
      setItemsError("saved_filter_name_required");
      return;
    }
    setItemsError("");
    try {
      const payload = {
        name: savedFilterName.trim(),
        query: {
          text: collectionQuery.text,
          brand: collectionQuery.brand,
          category: collectionQuery.category,
          sort_by: collectionQuery.sort_by,
        },
      };
      const resp = await fetch(`/api/profiles/${encodeURIComponent(activeProfile.id)}/saved-filters`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!resp.ok) {
        throw new Error("failed_to_save_filter");
      }
      const created = (await resp.json()) as { id: string; name: string; query?: Record<string, unknown> };
      setSavedFilters((current) => [...current, created]);
      setSavedFilterName("");
    } catch (e) {
      setItemsError(e instanceof Error ? e.message : "failed_to_save_filter");
    }
  }

  function applySavedFilter(filter: { query?: Record<string, unknown> }) {
    const q = filter.query || {};
    setCollectionQuery({
      text: String(q.text || ""),
      brand: String(q.brand || ""),
      category: String(q.category || ""),
      sort_by: String(q.sort_by || "date_added"),
    });
  }

  async function loadPhotos() {
    if (!selectedItemID) {
      setPhotosError("item_id_required");
      return;
    }
    setPhotosError("");
    try {
      const resp = await fetch(`/api/items/${encodeURIComponent(selectedItemID)}/photos`);
      if (!resp.ok) {
        throw new Error("failed_to_list_photos");
      }
      const data = (await resp.json()) as { photos?: Array<{ id: string; item_id: string; filename: string; is_primary?: boolean }> };
      setPhotos(data.photos || []);
    } catch (e) {
      setPhotosError(e instanceof Error ? e.message : "failed_to_list_photos");
    }
  }

  async function uploadPhoto() {
    if (!selectedItemID || !photoFile) {
      setPhotosError("item_and_file_required");
      return;
    }
    setPhotosError("");
    try {
      const form = new FormData();
      form.append("file", photoFile);
      const resp = await fetch(`/api/items/${encodeURIComponent(selectedItemID)}/photos`, {
        method: "POST",
        body: form,
      });
      if (!resp.ok) {
        throw new Error("failed_to_upload_photo");
      }
      await loadPhotos();
    } catch (e) {
      setPhotosError(e instanceof Error ? e.message : "failed_to_upload_photo");
    }
  }

  async function openCamera() {
    setCameraOpen(true);
    setCameraStatus("requesting_camera");
    setPhotosError("");
    try {
      if (!navigator.mediaDevices?.getUserMedia) {
        throw new Error("camera_not_supported");
      }
      const stream = await navigator.mediaDevices.getUserMedia({ video: true });
      stream.getTracks().forEach((track) => track.stop());
      setCameraStatus("camera_ready");
    } catch {
      setCameraStatus("camera_unavailable");
    }
  }

  async function setPrimaryPhoto(photoID: string) {
    if (!selectedItemID || !photoID) {
      return;
    }
    setPhotosError("");
    try {
      const resp = await fetch(`/api/items/${encodeURIComponent(selectedItemID)}/photos/${encodeURIComponent(photoID)}/primary`, {
        method: "PUT",
      });
      if (!resp.ok) {
        throw new Error("failed_to_set_primary_photo");
      }
      await loadPhotos();
    } catch (e) {
      setPhotosError(e instanceof Error ? e.message : "failed_to_set_primary_photo");
    }
  }

  async function deletePhoto(photoID: string) {
    if (!selectedItemID || !photoID) {
      return;
    }
    setPhotosError("");
    try {
      const resp = await fetch(`/api/items/${encodeURIComponent(selectedItemID)}/photos/${encodeURIComponent(photoID)}`, {
        method: "DELETE",
      });
      if (!resp.ok && resp.status !== 204) {
        throw new Error("failed_to_delete_photo");
      }
      await loadPhotos();
    } catch (e) {
      setPhotosError(e instanceof Error ? e.message : "failed_to_delete_photo");
    }
  }

  async function loadQuerySets() {
    setScannerError("");
    try {
      const resp = await fetch("/api/scanner/query-sets");
      if (!resp.ok) {
        throw new Error("failed_to_list_query_sets");
      }
      const data = (await resp.json()) as { query_sets?: ScannerQuerySetRecord[] };
      const listed = data.query_sets || [];
      setQuerySets(listed);
      if (listed.length > 0 && !selectedQuerySetID) {
        setSelectedQuerySetID(listed[0].id);
      }
    } catch (e) {
      setScannerError(e instanceof Error ? e.message : "failed_to_list_query_sets");
    }
  }

  async function createQuerySet(values: ScannerQuerySetValues) {
    setScannerError("");
    setQuerySetSubmitting(true);
    try {
      const payload = {
        name: values.name.trim() || "Query Set",
        keywords: values.keywords
          .split(",")
          .map((k) => k.trim())
          .filter(Boolean),
        exclusions: values.exclusions
          .split(",")
          .map((e) => e.trim())
          .filter(Boolean),
        max_price: values.max_price || 0,
        region: values.region || "",
        condition: values.condition || "",
        schedule_cron: values.schedule_cron || "",
        enabled: values.enabled,
        rate_limit_rps: values.rate_limit_rps,
        max_retry_count: values.max_retry_count,
      };
      const resp = await fetch("/api/scanner/query-sets", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!resp.ok) {
        throw new Error("failed_to_create_query_set");
      }
      const created = (await resp.json()) as ScannerQuerySetRecord;
      setQuerySets((current) => [...current, created]);
      if (!selectedQuerySetID) {
        setSelectedQuerySetID(created.id);
      }
      setEditingQuerySetID("");
    } catch (e) {
      setScannerError(e instanceof Error ? e.message : "failed_to_create_query_set");
    } finally {
      setQuerySetSubmitting(false);
    }
  }

  async function runScannerNow() {
    if (!selectedQuerySetID) {
      setScannerError("query_set_required");
      return;
    }
    setScannerError("");
    try {
      const resp = await fetch("/api/scanner/run", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ query_set_id: selectedQuerySetID }),
      });
      if (!resp.ok) {
        throw new Error("failed_to_run_scanner");
      }
      await loadCandidates();
    } catch (e) {
      setScannerError(e instanceof Error ? e.message : "failed_to_run_scanner");
    }
  }

  async function runScheduledScans() {
    setScannerError("");
    setScheduledRunStatus("");
    try {
      const resp = await fetch("/api/scanner/run/scheduled", { method: "POST", headers: { "Content-Type": "application/json" }, body: "{}" });
      if (!resp.ok) {
        throw new Error("failed_to_run_scheduled_scans");
      }
      setScheduledRunStatus("scheduled_scans_triggered");
    } catch (e) {
      setScannerError(e instanceof Error ? e.message : "failed_to_run_scheduled_scans");
    }
  }

  async function loadScannerFailures() {
    setScannerError("");
    try {
      const resp = await fetch("/api/scanner/failures");
      if (!resp.ok) {
        throw new Error("failed_to_load_scanner_failures");
      }
      const data = (await resp.json()) as { failures?: Array<{ id?: string; query_set_id?: string; reason?: string; attempts?: number; last_error_at?: string }> };
      setScannerFailures(data.failures || []);
    } catch (e) {
      setScannerError(e instanceof Error ? e.message : "failed_to_load_scanner_failures");
    }
  }

  async function retryScannerFailure(querySetID: string) {
    const trimmed = querySetID.trim();
    if (!trimmed) {
      setScannerError("missing_query_set_id");
      return;
    }
    setScannerError("");
    setScannerRetryStatus("");
    try {
      const resp = await fetch("/api/scanner/failures/retry", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ query_set_id: trimmed }),
      });
      if (!resp.ok) {
        throw new Error("failed_to_retry_scanner_failure");
      }
      setSelectedQuerySetID(trimmed);
      setScannerRetryStatus(`retry_started_for_${trimmed}`);
      await loadScannerFailures();
    } catch (e) {
      setScannerError(e instanceof Error ? e.message : "failed_to_retry_scanner_failure");
    }
  }

  async function loadProviderHealth(provider = "ebay") {
    setScannerError("");
    try {
      const resp = await fetch(`/api/provider/health?provider=${encodeURIComponent(provider)}`);
      if (!resp.ok) {
        throw new Error("failed_to_load_provider_health");
      }
      const data = (await resp.json()) as { provider?: string; state?: string; healthy?: boolean };
      setProviderHealth(data);
    } catch (e) {
      setScannerError(e instanceof Error ? e.message : "failed_to_load_provider_health");
    }
  }

  async function runMatchingNow() {
    setScannerError("");
    setMatchingRunStatus("");
    try {
      const resp = await fetch("/api/matching/run", { method: "POST", headers: { "Content-Type": "application/json" }, body: "{}" });
      if (!resp.ok) {
        throw new Error("failed_to_run_matching");
      }
      const data = (await resp.json()) as { matched?: number; suggested?: number; not_in_collection?: number; processed?: number };
      const processed = Number(data.processed || 0);
      setMatchingRunStatus(`matching_run_ok:${processed}`);
      await loadMatchingResults();
    } catch (e) {
      setScannerError(e instanceof Error ? e.message : "failed_to_run_matching");
    }
  }

  async function loadCandidates() {
    if (!selectedQuerySetID) {
      setScannerError("query_set_required");
      return;
    }
    setScannerError("");
    try {
      const resp = await fetch(`/api/scanner/candidates?query_set_id=${encodeURIComponent(selectedQuerySetID)}`);
      if (!resp.ok) {
        throw new Error("failed_to_list_candidates");
      }
      const data = (await resp.json()) as { candidates?: Array<{ id: string; title?: string; listing_id?: string; status?: string }> };
      setCandidates(data.candidates || []);
    } catch (e) {
      setScannerError(e instanceof Error ? e.message : "failed_to_list_candidates");
    }
  }

  async function discoveryAction(candidateID: string, actionType: string) {
    setScannerError("");
    try {
      const resp = await fetch("/api/discovery/action", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ candidate_id: candidateID, type: actionType }),
      });
      if (!resp.ok) {
        throw new Error("failed_to_apply_discovery_action");
      }
      await loadCandidates();
    } catch (e) {
      setScannerError(e instanceof Error ? e.message : "failed_to_apply_discovery_action");
    }
  }

  async function loadMatchingResults() {
    setScannerError("");
    try {
      const resp = await fetch("/api/matching/results");
      if (!resp.ok) {
        throw new Error("failed_to_list_matching_results");
      }
      const data = (await resp.json()) as {
        results?: Array<{ candidate_id?: string; state?: string; part_number?: string; item_id?: string }>;
      };
      setMatchingResults(data.results || []);
    } catch (e) {
      setScannerError(e instanceof Error ? e.message : "failed_to_list_matching_results");
    }
  }

  async function loadNotInCollection() {
    setScannerError("");
    try {
      const params = new URLSearchParams();
      if (notInCollectionFilter.query.trim()) {
        params.set("q", notInCollectionFilter.query.trim());
      }
      if (notInCollectionFilter.maxPrice.trim()) {
        params.set("price_max", notInCollectionFilter.maxPrice.trim());
      }
      if (notInCollectionFilter.dateFrom.trim()) {
        params.set("date_from", notInCollectionFilter.dateFrom.trim());
      }
      const resp = await fetch(`/api/discovery/not-in-collection?${params.toString()}`);
      if (!resp.ok) {
        throw new Error("failed_to_list_not_in_collection");
      }
      const data = (await resp.json()) as {
        items?: Array<{ candidate_id: string; title?: string; price?: number; url?: string; last_seen?: string }>;
      };
      setNotInCollectionItems(data.items || []);
    } catch (e) {
      setScannerError(e instanceof Error ? e.message : "failed_to_list_not_in_collection");
    }
  }

  async function notInCollectionAction(candidateID: string, actionType: "ignore" | "add_to_wishlist" | "track_price" | "create_item") {
    setScannerError("");
    try {
      const resp = await fetch("/api/discovery/action", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ candidate_id: candidateID, type: actionType }),
      });
      if (!resp.ok) {
        throw new Error("failed_to_apply_discovery_action");
      }
      await loadNotInCollection();
    } catch (e) {
      setScannerError(e instanceof Error ? e.message : "failed_to_apply_discovery_action");
    }
  }

  async function loadDashboard() {
    setInsightError("");
    setDashboardLoading(true);
    try {
      const resp = await fetch("/api/dashboard");
      if (!resp.ok) {
        throw new Error("failed_to_load_dashboard");
      }
      const data = (await resp.json()) as Record<string, unknown>;
      setDashboard(data);
    } catch (e) {
      setInsightError(e instanceof Error ? e.message : "failed_to_load_dashboard");
    } finally {
      setDashboardLoading(false);
    }
  }

  async function loadWishlist() {
    setInsightError("");
    try {
      const resp = await fetch("/api/wishlist");
      if (!resp.ok) {
        throw new Error("failed_to_load_wishlist");
      }
      const data = (await resp.json()) as { wishlist?: Array<{ id: string; item_id: string; target_price?: number; below_target_now?: boolean; priority?: string }> };
      setWishlist(data.wishlist || []);
    } catch (e) {
      setInsightError(e instanceof Error ? e.message : "failed_to_load_wishlist");
    }
  }

  async function createWishlistEntry() {
    setInsightError("");
    const itemID = wishlistDraft.item_id.trim() || selectedItemID;
    if (!itemID) {
      setInsightError("wishlist_item_id_required");
      return;
    }
    try {
      const payload = {
        item_id: itemID,
        target_price: Number(wishlistDraft.target_price || "0"),
        priority: wishlistDraft.priority || "normal",
        notes: wishlistDraft.notes || "",
      };
      const resp = await fetch("/api/wishlist", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!resp.ok) {
        throw new Error("failed_to_create_wishlist_entry");
      }
      await loadWishlist();
    } catch (e) {
      setInsightError(e instanceof Error ? e.message : "failed_to_create_wishlist_entry");
    }
  }

  async function deleteWishlistEntry(id: string) {
    setInsightError("");
    try {
      const resp = await fetch(`/api/wishlist?id=${encodeURIComponent(id)}`, { method: "DELETE" });
      if (!resp.ok && resp.status !== 204) {
        throw new Error("failed_to_delete_wishlist_entry");
      }
      await loadWishlist();
    } catch (e) {
      setInsightError(e instanceof Error ? e.message : "failed_to_delete_wishlist_entry");
    }
  }

  async function loadPricingGraph() {
    setInsightError("");
    if (!selectedItemID) {
      setInsightError("item_id_required_for_pricing");
      return;
    }
    try {
      const resp = await fetch(`/api/pricing/graph?item_id=${encodeURIComponent(selectedItemID)}`);
      if (!resp.ok) {
        throw new Error("failed_to_load_pricing_graph");
      }
      const data = (await resp.json()) as { points?: Array<{ day?: string; date?: string; price?: number; min?: number; median?: number; latest?: number }> };
      setPricingPoints(data.points || []);
    } catch (e) {
      setInsightError(e instanceof Error ? e.message : "failed_to_load_pricing_graph");
    }
  }

  async function loadPricingBySource() {
    setInsightError("");
    if (!selectedItemID) {
      setInsightError("item_id_required_for_pricing");
      return;
    }
    try {
      const resp = await fetch(`/api/pricing/by-source?item_id=${encodeURIComponent(selectedItemID)}`);
      if (!resp.ok) {
        throw new Error("failed_to_load_pricing_by_source");
      }
      const data = (await resp.json()) as {
        by_source?: Record<string, Array<{ snapshot_date?: string; min_price?: number; median_price?: number; latest_price?: number }>>;
      };
      setPricingBySource(data.by_source || {});
    } catch (e) {
      setInsightError(e instanceof Error ? e.message : "failed_to_load_pricing_by_source");
    }
  }

  async function trackPricingItem() {
    setInsightError("");
    setPricingTrackStatus("");
    if (!selectedItemID) {
      setInsightError("item_id_required_for_pricing_track");
      return;
    }
    try {
      const resp = await fetch("/api/pricing/track", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ item_id: selectedItemID }),
      });
      if (!resp.ok) {
        throw new Error("failed_to_track_pricing_item");
      }
      setPricingTrackStatus("pricing_track_enabled");
    } catch (e) {
      setInsightError(e instanceof Error ? e.message : "failed_to_track_pricing_item");
    }
  }

  async function loadPricingHistory() {
    if (!selectedItemID) {
      setInsightError("item_id_required_for_pricing_history");
      return;
    }
    setInsightError("");
    try {
      const resp = await fetch(`/api/pricing/history?item_id=${encodeURIComponent(selectedItemID)}`);
      if (!resp.ok) {
        throw new Error("failed_to_load_pricing_history");
      }
      const data = (await resp.json()) as { history?: Array<{ day?: string; date?: string; min?: number; median?: number; latest?: number }> };
      setPricingHistory(data.history || []);
    } catch (e) {
      setInsightError(e instanceof Error ? e.message : "failed_to_load_pricing_history");
    }
  }

  async function loadPricingStats() {
    if (!selectedItemID) {
      setInsightError("item_id_required_for_pricing_stats");
      return;
    }
    setInsightError("");
    try {
      const resp = await fetch(`/api/pricing/stats?item_id=${encodeURIComponent(selectedItemID)}`);
      if (!resp.ok) {
        throw new Error("failed_to_load_pricing_stats");
      }
      const data = (await resp.json()) as Record<string, unknown>;
      setPricingStats(data);
    } catch (e) {
      setInsightError(e instanceof Error ? e.message : "failed_to_load_pricing_stats");
    }
  }

  async function loadPricingTrend() {
    if (!selectedItemID) {
      setInsightError("item_id_required_for_pricing_trend");
      return;
    }
    setInsightError("");
    try {
      const resp = await fetch(`/api/pricing/trend?item_id=${encodeURIComponent(selectedItemID)}`);
      if (!resp.ok) {
        throw new Error("failed_to_load_pricing_trend");
      }
      const data = (await resp.json()) as Record<string, unknown>;
      setPricingTrend(data);
    } catch (e) {
      setInsightError(e instanceof Error ? e.message : "failed_to_load_pricing_trend");
    }
  }

  async function loadWishlistHits() {
    setInsightError("");
    try {
      const resp = await fetch("/api/wishlist/hits");
      if (!resp.ok) {
        throw new Error("failed_to_load_wishlist_hits");
      }
      const data = (await resp.json()) as { hits?: Array<{ item_id?: string; listing_id?: string; title?: string; price?: number }> };
      setWishlistHits(data.hits || []);
    } catch (e) {
      setInsightError(e instanceof Error ? e.message : "failed_to_load_wishlist_hits");
    }
  }

  async function runPricingSnapshot() {
    setInsightError("");
    setSnapshotStatus("");
    try {
      const resp = await fetch("/api/pricing/snapshot/run", { method: "POST", headers: { "Content-Type": "application/json" }, body: "{}" });
      if (!resp.ok) {
        throw new Error("failed_to_run_pricing_snapshot");
      }
      setSnapshotStatus("pricing_snapshot_completed");
    } catch (e) {
      setInsightError(e instanceof Error ? e.message : "failed_to_run_pricing_snapshot");
    }
  }

  async function loadAdminStatus() {
    if (!activeProfile?.id) {
      return;
    }
    setAdminError("");
    try {
      await refreshLicenseStatus(activeProfile.id);

      const logsResp = await fetch("/api/logs/activity?limit=10");
      if (!logsResp.ok) {
        throw new Error("failed_to_load_activity_logs");
      }
      const logs = (await logsResp.json()) as { activity?: Array<Record<string, unknown>>; logs?: Array<Record<string, unknown>> };
      const entries = logs.logs || logs.activity || [];
      setActivityLogs(entries);
      setLogCount(entries.length);
    } catch (e) {
      setAdminError(e instanceof Error ? e.message : "failed_to_load_admin_status");
    }
  }

  async function loadRuntimeDiagnostics() {
    setAdminError("");
    try {
      const [runtimeResp, recoveryResp] = await Promise.all([fetch("/api/runtime"), fetch("/api/runtime/recovery")]);
      if (!runtimeResp.ok || !recoveryResp.ok) {
        throw new Error("failed_to_load_runtime_diagnostics");
      }
      const runtime = (await runtimeResp.json()) as { update_channel?: string; update_public_key_configured?: boolean };
      const recovery = (await recoveryResp.json()) as { recovery_required?: boolean };
      setRuntimeDiagnostics(runtime);
      setRecoveryDiagnostics(recovery);
      setSettingsStatus("runtime_diagnostics_loaded");
    } catch (e) {
      setRuntimeDiagnostics(null);
      setRecoveryDiagnostics(null);
      setAdminError(e instanceof Error ? e.message : "failed_to_load_runtime_diagnostics");
    }
  }

  async function refreshLicenseStatus(profileID = activeProfile?.id || "") {
    if (!profileID) {
      return;
    }
    const licenseResp = await fetch(`/api/license/status?profile_id=${encodeURIComponent(profileID)}`);
    if (!licenseResp.ok) {
      throw new Error("failed_to_load_license_status");
    }
    const license = (await licenseResp.json()) as { state?: string; tier?: string; features?: string[]; expires_at?: string };
    setLicenseStatus(license);
  }

  async function loadProfileLicense() {
    if (!activeProfile?.id) {
      return;
    }
    setAdminError("");
    try {
      const resp = await fetch(`/api/profiles/${encodeURIComponent(activeProfile.id)}/license`);
      if (!resp.ok) {
        throw new Error("failed_to_get_profile_license");
      }
      const data = (await resp.json()) as { license_json?: string };
      setProfileLicenseJSON(data.license_json || "");
      setSettingsStatus("license_profile_loaded");
    } catch (e) {
      setAdminError(e instanceof Error ? e.message : "failed_to_get_profile_license");
    }
  }

  async function saveProfileLicense() {
    if (!activeProfile?.id) {
      return;
    }
    setAdminError("");
    try {
      const resp = await fetch(`/api/profiles/${encodeURIComponent(activeProfile.id)}/license`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ license_json: profileLicenseJSON }),
      });
      if (!resp.ok) {
        throw new Error("failed_to_put_profile_license");
      }
      setSettingsStatus("license_profile_saved");
    } catch (e) {
      setAdminError(e instanceof Error ? e.message : "failed_to_put_profile_license");
    }
  }

  async function importLicenseFile() {
    if (!activeProfile?.id) {
      return;
    }
    setAdminError("");
    setLicenseImportStatus("");
    if (!licenseImportDraft.payload_base64 || !licenseImportDraft.signature_base64) {
      setAdminError("license_file_fields_required");
      return;
    }
    try {
      const resp = await fetch("/api/license/import", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          profile_id: activeProfile.id,
          license: {
            payload_base64: licenseImportDraft.payload_base64,
            signature_base64: licenseImportDraft.signature_base64,
          },
        }),
      });
      if (!resp.ok) {
        throw new Error("failed_to_import_license");
      }
      setLicenseImportStatus("license_imported");
      await refreshLicenseStatus(activeProfile.id);
      setSettingsStatus("license_imported");
    } catch (e) {
      setAdminError(e instanceof Error ? e.message : "failed_to_import_license");
    }
  }

  async function loadProfileSettings() {
    if (!activeProfile?.id) {
      return;
    }
    setAdminError("");
    try {
      const resp = await fetch(`/api/profiles/${encodeURIComponent(activeProfile.id)}/settings`);
      if (!resp.ok) {
        throw new Error("failed_to_get_settings");
      }
      const data = (await resp.json()) as { settings?: Record<string, string> };
      const settings = data.settings || {};
      setProfileSettingsInitial({
        scanner_schedule: settings.scanner_schedule || "",
        backup_frequency: (settings.backup_frequency as ProfileSettingsValues["backup_frequency"]) || "daily",
        db_path: settings["storage.db_path"] || "",
        update_channel: (settings.update_channel as ProfileSettingsValues["update_channel"]) || "stable",
      });
      setDebugModeEnabled(String(settings.debug_mode || "").toLowerCase() === "true");
    } catch (e) {
      setAdminError(e instanceof Error ? e.message : "failed_to_get_settings");
    }
  }

  async function toggleDebugMode(nextEnabled: boolean) {
    if (!activeProfile?.id) {
      return;
    }
    setAdminError("");
    try {
      const resp = await fetch("/api/logs/debug", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ profile_id: activeProfile.id, enabled: nextEnabled }),
      });
      if (!resp.ok) {
        throw new Error("failed_to_toggle_debug_mode");
      }
      setDebugModeEnabled(nextEnabled);
      setSettingsStatus(nextEnabled ? "debug_mode_enabled" : "debug_mode_disabled");
    } catch (e) {
      setAdminError(e instanceof Error ? e.message : "failed_to_toggle_debug_mode");
    }
  }

  async function saveProfileSettings(values: ProfileSettingsValues) {
    if (!activeProfile?.id) {
      return;
    }
    setAdminError("");
    try {
      const payload = {
        settings: {
          scanner_schedule: values.scanner_schedule,
          backup_frequency: values.backup_frequency,
          "storage.db_path": values.db_path,
          update_channel: values.update_channel,
        },
      };
      const resp = await fetch(`/api/profiles/${encodeURIComponent(activeProfile.id)}/settings`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!resp.ok) {
        throw new Error("failed_to_update_settings");
      }
      setSettingsStatus("settings_saved");
    } catch (e) {
      setAdminError(e instanceof Error ? e.message : "failed_to_update_settings");
    }
  }

  async function saveSecrets(values: SecretsValues) {
    if (!activeProfile?.id) {
      return;
    }
    setAdminError("");
    try {
      const writes = [
        { key: "openai_api_key", value: values.openai_api_key },
        { key: "ebay_app_id", value: values.ebay_app_id },
        { key: "ebay_auth_token", value: values.ebay_auth_token },
      ];
      for (const write of writes) {
        const resp = await fetch(`/api/profiles/${encodeURIComponent(activeProfile.id)}/secrets`, {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(write),
        });
        if (!resp.ok) {
          throw new Error("failed_to_put_secret");
        }
      }
      setSecretsInitial(values);
      setSettingsStatus("secrets_saved");
    } catch (e) {
      setAdminError(e instanceof Error ? e.message : "failed_to_put_secret");
    }
  }

  async function resetIgnoreRules() {
    setAdminError("");
    try {
      const resp = await fetch("/api/settings/reset-ignore-rules", { method: "POST", headers: { "Content-Type": "application/json" }, body: "{}" });
      if (!resp.ok) {
        throw new Error("failed_to_reset_ignore_rules");
      }
      setSettingsStatus("ignore_rules_reset_ok");
    } catch (e) {
      setAdminError(e instanceof Error ? e.message : "failed_to_reset_ignore_rules");
    }
  }

  async function rebuildThumbnails() {
    if (!selectedItemID) {
      setAdminError("item_id_required_for_rebuild");
      return;
    }
    setAdminError("");
    try {
      const resp = await fetch(`/api/items/${encodeURIComponent(selectedItemID)}/photos-rebuild`, { method: "POST", headers: { "Content-Type": "application/json" }, body: "{}" });
      if (!resp.ok) {
        throw new Error("failed_to_rebuild_thumbnails");
      }
      setSettingsStatus("thumbnails_rebuild_ok");
    } catch (e) {
      setAdminError(e instanceof Error ? e.message : "failed_to_rebuild_thumbnails");
    }
  }

  async function runDataMaintenance(path: string, label: string) {
    setAdminError("");
    try {
      const resp = await fetch(path, { method: "POST", headers: { "Content-Type": "application/json" }, body: "{}" });
      if (!resp.ok) {
        throw new Error(`failed_${label}`);
      }
      setSettingsStatus(`${label}_ok`);
    } catch (e) {
      setAdminError(e instanceof Error ? e.message : `failed_${label}`);
    }
  }

  function parseBackupEntry(path: string) {
    const normalized = path.replaceAll("\\", "/");
    const name = normalized.split("/").pop() || path;
    const m = name.match(/cabinet-backup-(\d{4})(\d{2})(\d{2})-(\d{2})(\d{2})(\d{2})/);
    const timestampLabel = m ? `${m[1]}-${m[2]}-${m[3]} ${m[4]}:${m[5]}:${m[6]} UTC` : "unknown time";
    return { path, name, timestampLabel };
  }

  async function loadBackups() {
    setAdminError("");
    try {
      const resp = await fetch("/api/backup/list");
      if (!resp.ok) {
        throw new Error("failed_to_list_backups");
      }
      const data = (await resp.json()) as { backups?: string[] };
      const next = (data.backups || []).map(parseBackupEntry);
      setBackupEntries(next);
      setSelectedBackupPath(next[0]?.path || "");
      setConfirmRestore(false);
      setSettingsStatus("backup_list_loaded");
    } catch (e) {
      setAdminError(e instanceof Error ? e.message : "failed_to_list_backups");
    }
  }

  async function restoreSelectedBackup() {
    if (!selectedBackupPath) {
      setAdminError("backup_selection_required");
      return;
    }
    if (!confirmRestore) {
      setAdminError("restore_confirmation_required");
      return;
    }
    setAdminError("");
    try {
      const resp = await fetch("/api/backup/restore", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ backup_path: selectedBackupPath }),
      });
      if (!resp.ok) {
        throw new Error("failed_to_restore_backup");
      }
      setSettingsStatus("backup_restored");
      setConfirmRestore(false);
    } catch (e) {
      setAdminError(e instanceof Error ? e.message : "failed_to_restore_backup");
    }
  }

  async function exportText(path: string, label: string) {
    try {
      const resp = await fetch(path);
      if (!resp.ok) {
        throw new Error(`failed_export_${label}`);
      }
      const text = await resp.text();
      setExportBytes(text.length);
      setSettingsStatus(`${label}_export_ok`);
    } catch (e) {
      setAdminError(e instanceof Error ? e.message : `failed_export_${label}`);
    }
  }

  async function dataImportDryRun(args: { format: "json" | "csv"; payload: string; mapping: Record<string, string> }) {
    setAdminError("");
    setSettingsStatus("");
    try {
      if (args.format === "json") {
        const resp = await fetch("/api/data/import/json/dry-run", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ snapshot: JSON.parse(args.payload) }),
        });
        if (!resp.ok) {
          throw new Error("failed_to_dry_run_import");
        }
        const data = (await resp.json()) as Record<string, unknown>;
        setSettingsStatus("import_dry_run_ok");
        return data;
      }
      const resp = await fetch("/api/data/import/csv/dry-run", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ csv_import: { csv: args.payload, mapping: args.mapping } }),
      });
      if (!resp.ok) {
        throw new Error("failed_to_dry_run_import");
      }
      const data = (await resp.json()) as Record<string, unknown>;
      setSettingsStatus("import_dry_run_ok");
      return data;
    } catch (e) {
      setAdminError(e instanceof Error ? e.message : "failed_to_dry_run_import");
      throw e;
    }
  }

  async function dataImportApply(args: {
    format: "json" | "csv";
    payload: string;
    mapping: Record<string, string>;
    options: { default_action: string; overrides: Record<string, string> };
  }) {
    setAdminError("");
    setSettingsStatus("");
    try {
      if (args.format === "json") {
        const resp = await fetch("/api/data/import/json/apply", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ snapshot: JSON.parse(args.payload), options: args.options }),
        });
        if (!resp.ok) {
          throw new Error("failed_to_apply_import");
        }
      } else {
        const resp = await fetch("/api/data/import/csv/apply", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ csv_import: { csv: args.payload, mapping: args.mapping }, options: args.options }),
        });
        if (!resp.ok) {
          throw new Error("failed_to_apply_import");
        }
      }
      setSettingsStatus("import_apply_ok");
    } catch (e) {
      setAdminError(e instanceof Error ? e.message : "failed_to_apply_import");
      throw e;
    }
  }

  async function dataExportRun(args: { format: "json" | "csv"; scope: "full" | "items" }) {
    if (args.format === "json") {
      await exportText("/api/data/export/json", "json");
      return;
    }
    await exportText("/api/data/export/csv/items", "csv_items");
  }

  async function loadItemBarcodes() {
    if (!selectedItemID) {
      setBarcodeError("item_id_required");
      return;
    }
    setBarcodeError("");
    try {
      const resp = await fetch(`/api/items/${encodeURIComponent(selectedItemID)}/barcodes`);
      if (!resp.ok) {
        throw new Error("failed_to_list_barcodes");
      }
      const data = (await resp.json()) as { barcodes?: Array<{ id?: string; barcode: string }> };
      setBarcodes(data.barcodes || []);
      if ((data.barcodes || []).length > 0 && !barcodeInput) {
        setBarcodeInput((data.barcodes || [])[0].barcode);
      }
    } catch (e) {
      setBarcodeError(e instanceof Error ? e.message : "failed_to_list_barcodes");
    }
  }

  async function addBarcode() {
    if (!selectedItemID || !barcodeInput.trim()) {
      setBarcodeError("item_and_barcode_required");
      return;
    }
    setBarcodeError("");
    try {
      const resp = await fetch(`/api/items/${encodeURIComponent(selectedItemID)}/barcodes`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ barcode: barcodeInput.trim() }),
      });
      if (!resp.ok) {
        throw new Error("failed_to_add_barcode");
      }
      await loadItemBarcodes();
    } catch (e) {
      setBarcodeError(e instanceof Error ? e.message : "failed_to_add_barcode");
    }
  }

  async function lookupBarcode() {
    if (!barcodeInput.trim()) {
      setBarcodeError("barcode_required");
      return;
    }
    setBarcodeError("");
    try {
      const resp = await fetch(`/api/barcodes/${encodeURIComponent(barcodeInput.trim())}`);
      if (!resp.ok) {
        throw new Error("failed_to_lookup_barcode");
      }
      const data = (await resp.json()) as { matches?: Array<{ item_id?: string; part_number?: string }> };
      setBarcodeLookupMatches(data.matches || []);
    } catch (e) {
      setBarcodeError(e instanceof Error ? e.message : "failed_to_lookup_barcode");
    }
  }

  async function externalBarcodeSearch() {
    if (!barcodeInput.trim()) {
      setBarcodeError("barcode_required");
      return;
    }
    setBarcodeError("");
    try {
      const resp = await fetch(
        `/api/barcodes/${encodeURIComponent(barcodeInput.trim())}/external-search?source=ebay&region=US`,
      );
      if (!resp.ok) {
        throw new Error("failed_to_build_external_search");
      }
      const data = (await resp.json()) as { url?: string };
      setBarcodeExternalURL(data.url || "");
    } catch (e) {
      setBarcodeError(e instanceof Error ? e.message : "failed_to_build_external_search");
    }
  }

  async function toggleAI(enabled: boolean) {
    if (!activeProfile?.id) {
      return;
    }
    setAiError("");
    try {
      const resp = await fetch("/api/ai/toggle", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ profile_id: activeProfile.id, enabled }),
      });
      if (!resp.ok) {
        throw new Error("failed_to_toggle_ai");
      }
      setAiEnabled(enabled);
    } catch (e) {
      setAiError(e instanceof Error ? e.message : "failed_to_toggle_ai");
    }
  }

  async function testAI() {
    if (!activeProfile?.id) {
      return;
    }
    setAiError("");
    try {
      const resp = await fetch("/api/ai/test", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ profile_id: activeProfile.id }),
      });
      if (!resp.ok) {
        throw new Error("failed_ai_test");
      }
      setAiSuggestion({ title: "AI connectivity ok", confidence: 1 });
    } catch (e) {
      setAiError(e instanceof Error ? e.message : "failed_ai_test");
    }
  }

  async function suggestFromTitle(input?: string) {
    const title = (input ?? aiTitleInput).trim();
    if (!activeProfile?.id || !title) {
      setAiError("profile_and_title_required");
      return;
    }
    setAiError("");
    setAiLastAction("title");
    setAiTitleInput(title);
    try {
      const resp = await fetch("/api/ai/suggest/title", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ profile_id: activeProfile.id, title }),
      });
      if (!resp.ok) {
        throw new Error("failed_ai_suggest_title");
      }
      const suggestion = (await resp.json()) as { title?: string; confidence?: number; [key: string]: unknown };
      setAiSuggestion(suggestion);
    } catch (e) {
      setAiError(e instanceof Error ? e.message : "failed_ai_suggest_title");
    }
  }

  async function suggestFromPhoto(input?: string) {
    const photo = (input ?? aiPhotoURL).trim();
    if (!activeProfile?.id || !photo) {
      setAiError("profile_and_photo_required");
      return;
    }
    setAiError("");
    setAiLastAction("photo");
    setAiPhotoURL(photo);
    try {
      const resp = await fetch("/api/ai/suggest/photo", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ profile_id: activeProfile.id, image_url: photo }),
      });
      if (!resp.ok) {
        throw new Error("failed_ai_suggest_photo");
      }
      const suggestion = (await resp.json()) as { title?: string; confidence?: number; [key: string]: unknown };
      setAiSuggestion(suggestion);
    } catch (e) {
      setAiError(e instanceof Error ? e.message : "failed_ai_suggest_photo");
    }
  }

  function applySuggestion() {
    if (!aiSuggestion?.title) {
      return;
    }
    setCollectionFormSeed((current) => ({ ...current, title: String(aiSuggestion.title || "") }));
    setCollectionFormKey((current) => current + 1);
  }

  async function retryLastAIAction() {
    if (aiLastAction === "title") {
      await suggestFromTitle(aiTitleInput);
      return;
    }
    if (aiLastAction === "photo") {
      await suggestFromPhoto(aiPhotoURL);
    }
  }

  async function beginWebAuthnRegistration() {
    if (!activeProfile?.id) {
      return;
    }
    setError("");
    try {
      const data = await postJSON("/api/auth/webauthn/register/begin", { profile_id: activeProfile.id });
      const sid = String(data.session_id || "");
      setAuthSessionID(sid);
      setAuthStatus("registration_begin_ready");
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_begin_registration");
    }
  }

  async function finishWebAuthnRegistration() {
    if (!authSessionID) {
      return;
    }
    setError("");
    try {
      const parsed = JSON.parse(credentialJSON || "{}");
      await postJSON("/api/auth/webauthn/register/finish", { session_id: authSessionID, credential: parsed });
      setAuthStatus("registration_finished");
      if (activeProfile?.id) {
        setOnboardingIdentityComplete(true);
        localStorage.setItem(onboardingIdentityPreferenceKey(activeProfile.id), "1");
      }
      await seedOnboardingSampleData();
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_finish_registration");
    }
  }

  async function beginWebAuthnLogin() {
    if (!activeProfile?.id) {
      return;
    }
    setError("");
    try {
      const data = await postJSON("/api/auth/webauthn/login/begin", { profile_id: activeProfile.id });
      const sid = String(data.session_id || "");
      setAuthSessionID(sid);
      setAuthStatus("login_begin_ready");
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_begin_login");
    }
  }

  async function finishWebAuthnLogin() {
    if (!authSessionID) {
      return;
    }
    setError("");
    try {
      const parsed = JSON.parse(credentialJSON || "{}");
      const data = await postJSON("/api/auth/webauthn/login/finish", { session_id: authSessionID, credential: parsed });
      const token = String(data.session_token || "");
      if (token) {
        setSessionToken(token);
      }
      setAuthStatus("login_finished");
      if (activeProfile?.id) {
        setOnboardingIdentityComplete(true);
        localStorage.setItem(onboardingIdentityPreferenceKey(activeProfile.id), "1");
      }
      await seedOnboardingSampleData();
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_finish_login");
    }
  }

  async function seedOnboardingSampleData() {
    setOnboardingStatus("");
    try {
      const data = await postJSON("/api/onboarding/sample-data", {});
      const createdItems = Number(data.created_items || 0);
      if (createdItems > 0) {
        setOnboardingStatus(`Onboarding sample data loaded (${createdItems} starter items).`);
      } else {
        setOnboardingStatus("Onboarding sample data already available.");
      }
      await loadItems();
    } catch (e) {
      setOnboardingStatus("Onboarding sample data failed to load.");
      setError(e instanceof Error ? e.message : "failed_to_seed_onboarding_sample_data");
    }
  }

  async function completeIdentity() {
    if (!activeProfile?.id) {
      return;
    }
    setStarterIdentityBusy(true);
    setError("");
    try {
      if (requiresRegistration) {
        const start = await postJSON("/api/auth/webauthn/register/begin", { profile_id: activeProfile.id });
        const sid = String(start.session_id || "");
        setAuthSessionID(sid);
        await postJSON("/api/auth/webauthn/register/finish", { session_id: sid, credential: {} });
        setAuthStatus("registration_finished");
        setOnboardingIdentityComplete(true);
        localStorage.setItem(onboardingIdentityPreferenceKey(activeProfile.id), "1");
        await seedOnboardingSampleData();
      } else {
        const start = await postJSON("/api/auth/webauthn/login/begin", { profile_id: activeProfile.id });
        const sid = String(start.session_id || "");
        setAuthSessionID(sid);
        const done = await postJSON("/api/auth/webauthn/login/finish", { session_id: sid, credential: {} });
        const token = String(done.session_token || "");
        if (token) {
          setSessionToken(token);
        }
        setAuthStatus("login_finished");
        setOnboardingIdentityComplete(true);
        localStorage.setItem(onboardingIdentityPreferenceKey(activeProfile.id), "1");
        await seedOnboardingSampleData();
      }
    } catch (e) {
      setOnboardingIdentityComplete(false);
      localStorage.setItem(onboardingIdentityPreferenceKey(activeProfile.id), "0");
      setError(e instanceof Error ? e.message : "failed_to_complete_identity");
    } finally {
      setStarterIdentityBusy(false);
    }
  }

  async function saveRecoveryPassphrase(passphrase?: string) {
    if (!activeProfile?.id) {
      return;
    }
    const value = (passphrase ?? recoveryPassphrase).trim();
    if (!value) {
      setError("recovery_passphrase_required");
      return;
    }
    setError("");
    try {
      await postJSON("/api/auth/recovery/passphrase", { profile_id: activeProfile.id, passphrase: value });
      setAuthStatus("recovery_passphrase_set");
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_set_recovery_passphrase");
    }
  }

  async function beginRecoveryReset() {
    if (!activeProfile?.id || !recoveryPassphrase) {
      return;
    }
    setError("");
    try {
      const data = await postJSON("/api/auth/recovery/reset/begin", {
        profile_id: activeProfile.id,
        passphrase: recoveryPassphrase,
      });
      setAuthSessionID(String(data.session_id || ""));
      setAuthStatus("recovery_begin_ready");
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_begin_recovery_reset");
    }
  }

  async function validateSession(token?: string) {
    const value = (token ?? sessionToken).trim();
    if (!value) {
      return;
    }
    setError("");
    try {
      await postJSON("/api/auth/session/validate", { session_token: value });
      setAuthStatus("session_valid");
      if (activeProfile?.id) {
        setOnboardingIdentityComplete(true);
        localStorage.setItem(onboardingIdentityPreferenceKey(activeProfile.id), "1");
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : "session_invalid_or_locked");
    }
  }

  async function lockSession(token?: string) {
    const value = (token ?? sessionToken).trim();
    if (!value) {
      return;
    }
    setError("");
    try {
      await postJSON("/api/auth/session/lock", { session_token: value });
      setAuthStatus("session_locked");
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_lock_session");
    }
  }

  const brands = Array.from(new Set(items.map((item) => item.brand || "Unknown")));
  const categoriesForBrand = Array.from(
    new Set(items.filter((item) => (columnBrand ? (item.brand || "Unknown") === columnBrand : true)).map((item) => item.category || "Unknown")),
  );
  const seriesForCategory = Array.from(
    new Set(
      items
        .filter((item) => {
          const brandOK = columnBrand ? (item.brand || "Unknown") === columnBrand : true;
          const categoryOK = columnCategory ? (item.category || "Unknown") === columnCategory : true;
          return brandOK && categoryOK;
        })
        .map((item) => (item as { series?: string }).series || "Unknown"),
    ),
  );
  const columnItems = items.filter((item) => {
    const brandOK = columnBrand ? (item.brand || "Unknown") === columnBrand : true;
    const categoryOK = columnCategory ? (item.category || "Unknown") === columnCategory : true;
    const seriesOK = columnSeries ? ((item as { series?: string }).series || "Unknown") === columnSeries : true;
    return brandOK && categoryOK && seriesOK;
  });
  const matchedCount = matchingResults.filter((result) => result.state === "matched").length;
  const suggestedCount = matchingResults.filter((result) => result.state === "suggested").length;
  const notInCollectionCount = matchingResults.filter((result) => result.state === "not_in_collection").length;
  const editingQuerySet = querySets.find((querySet) => querySet.id === editingQuerySetID);
  const querySetInitialValues: Partial<ScannerQuerySetValues> | undefined = editingQuerySet
    ? {
        name: editingQuerySet.name || "",
        keywords: (editingQuerySet.keywords || []).join(", "),
        exclusions: (editingQuerySet.exclusions || []).join(", "),
        max_price: editingQuerySet.max_price,
        region: editingQuerySet.region || "US",
        condition: editingQuerySet.condition || "",
        schedule_cron: editingQuerySet.schedule_cron || "",
        enabled: editingQuerySet.enabled ?? true,
        rate_limit_rps: editingQuerySet.rate_limit_rps ?? 2,
        max_retry_count: editingQuerySet.max_retry_count ?? 1,
      }
    : undefined;
  const showAdvancedWorkspace = Boolean(activeProfile && advancedWorkspace);
  const navEnabled = showAdvancedWorkspace;
  function selectScreen(screen: TopLevelScreen) {
    if (!navEnabled) {
      return;
    }
    setActiveScreen(screen);
  }
  function openWorkspaceFromDashboard(screen: Exclude<TopLevelScreen, "all">) {
    selectScreen(screen);
    if (screen === "pricing") {
      void loadWishlist();
    }
  }
  const navLinks = (
    <>
      <button
        type="button"
        className={`cabinet-nav-link${activeScreen === "dashboard" ? " cabinet-nav-link-active" : ""}`}
        aria-current={activeScreen === "dashboard" ? "page" : undefined}
        onClick={() => selectScreen("dashboard")}
        disabled={!navEnabled}
      >
        Dashboard
      </button>
      <button
        type="button"
        className={`cabinet-nav-link${activeScreen === "collection" ? " cabinet-nav-link-active" : ""}`}
        aria-current={activeScreen === "collection" ? "page" : undefined}
        onClick={() => selectScreen("collection")}
        disabled={!navEnabled}
      >
        Collection
      </button>
      <button
        type="button"
        className={`cabinet-nav-link${activeScreen === "scanner" ? " cabinet-nav-link-active" : ""}`}
        aria-current={activeScreen === "scanner" ? "page" : undefined}
        onClick={() => selectScreen("scanner")}
        disabled={!navEnabled}
      >
        Scanner
      </button>
      <button
        type="button"
        className={`cabinet-nav-link${activeScreen === "pricing" ? " cabinet-nav-link-active" : ""}`}
        aria-current={activeScreen === "pricing" ? "page" : undefined}
        onClick={() => selectScreen("pricing")}
        disabled={!navEnabled}
      >
        Pricing
      </button>
      <button
        type="button"
        className={`cabinet-nav-link${activeScreen === "settings" ? " cabinet-nav-link-active" : ""}`}
        aria-current={activeScreen === "settings" ? "page" : undefined}
        onClick={() => selectScreen("settings")}
        disabled={!navEnabled}
      >
        Settings
      </button>
    </>
  );

  return (
    <main data-testid="app-shell" className="cabinet-shell">
      <aside className="cabinet-sidebar">
        <h1>Cabinet</h1>
        <nav>{navLinks}</nav>
      </aside>
      {mobileNavOpen ? (
        <>
          <button
            type="button"
            className="cabinet-drawer-backdrop"
            aria-label="Close navigation menu backdrop"
            onClick={() => setMobileNavOpen(false)}
          />
          <div
            id="cabinet-mobile-nav"
            ref={drawerRef}
            className="cabinet-drawer"
            role="dialog"
            aria-modal="true"
            aria-label="Navigation Menu"
            tabIndex={-1}
          >
            <div className="cabinet-drawer-head">
              <h2>Navigation</h2>
              <button type="button" onClick={() => setMobileNavOpen(false)}>
                Close
              </button>
            </div>
            <nav onClick={() => setMobileNavOpen(false)}>{navLinks}</nav>
          </div>
        </>
      ) : null}
      <section className="cabinet-content">
        <header className="cabinet-topbar">
          <div className="cabinet-topbar-left">
            <button
              ref={drawerTriggerRef}
              type="button"
              className="cabinet-nav-toggle"
              aria-controls="cabinet-mobile-nav"
              aria-expanded={mobileNavOpen}
              aria-label="Open navigation menu"
              onClick={() => setMobileNavOpen(true)}
            >
              Menu
            </button>
            <strong>Runtime connected. UI foundation active.</strong>
          </div>
          <button
            id="theme-toggle"
            type="button"
            onClick={() => setTheme((current) => (current === "dark" ? "light" : "dark"))}
          >
            Toggle Theme
          </button>
        </header>
        <section className="cabinet-card">
          <h2>Cabinet Frontend Foundation</h2>
          <p>Local-first onboarding, WebAuthn auth, collection workflows, and automated E2E coverage are active.</p>
          {loading ? <p>Loading profiles...</p> : null}
          {!loading && profiles.length === 0 ? (
            <div>
              <p>No local profiles found yet. Create your first local profile to initialize Cabinet on this device.</p>
              <button type="button" onClick={createFirstProfile}>
                Create First Profile
              </button>
            </div>
          ) : null}
          {!loading && profiles.length > 0 ? (
            <div>
              <p>Select a profile to continue:</p>
              <ul>
                {profiles.map((p) => (
                  <li key={p.id}>
                    <span>{p.name}</span>{" "}
                    <button type="button" onClick={() => activateProfile(p.id)}>
                      Use {p.name}
                    </button>
                  </li>
                ))}
              </ul>
            </div>
          ) : null}
          {activeProfile ? <p>Active profile: {activeProfile.name}</p> : null}
          {profileStorage ? (
            <div>
              <p>Database: {profileStorage.db_path || "not set"}</p>
              <p>Media: {profileStorage.media_dir || "not set"}</p>
            </div>
          ) : null}
          {activeProfile ? (
            <div>
              <h3>Authentication</h3>
              <p>Requires registration: {requiresRegistration === null ? "unknown" : String(requiresRegistration)}</p>
              <details>
                <summary>Advanced Identity Tools</summary>
                <div>
                  <button type="button" onClick={beginWebAuthnRegistration}>
                    Begin WebAuthn Registration
                  </button>{" "}
                  <button type="button" onClick={finishWebAuthnRegistration}>
                    Finish Registration
                  </button>{" "}
                  <button type="button" onClick={beginWebAuthnLogin}>
                    Begin WebAuthn Login
                  </button>{" "}
                  <button type="button" onClick={finishWebAuthnLogin}>
                    Finish Login
                  </button>
                </div>
                <p>Auth session: {authSessionID || "none"}</p>
                <textarea
                  value={credentialJSON}
                  onChange={(e) => setCredentialJSON(e.target.value)}
                  rows={4}
                  cols={60}
                  aria-label="Credential JSON"
                />
                <RecoveryPassphraseForm
                  isSubmitting={recoverySubmitting}
                  onSubmit={async ({ passphrase }) => {
                    setRecoverySubmitting(true);
                    setRecoveryPassphrase(passphrase);
                    try {
                      await saveRecoveryPassphrase(passphrase);
                    } finally {
                      setRecoverySubmitting(false);
                    }
                  }}
                />
                <button type="button" onClick={beginRecoveryReset}>
                  Begin Recovery Reset
                </button>
                <SessionTokenForm
                  onValidate={async ({ token }) => {
                    setSessionToken(token);
                    await validateSession(token);
                  }}
                  onLock={async ({ token }) => {
                    setSessionToken(token);
                    await lockSession(token);
                  }}
                />
              </details>
              <p>Auth status: {authStatus || "idle"}</p>
              {onboardingStatus ? <p>{onboardingStatus}</p> : null}
            </div>
          ) : null}
          {activeProfile && !showAdvancedWorkspace ? (
            <div>
              <h3>Starter Onboarding Wizard</h3>
              <p>
                Step {onboardingStep} of {ONBOARDING_STEPS.length}
              </p>
              <p>Current step: {ONBOARDING_STEPS[onboardingStep - 1]}</p>
              {onboardingSetupPath ? <p>Setup path: {onboardingSetupPath}</p> : null}
              {onboardingStep === 1 ? (
                <div>
                  <p>Choose how you want to begin. Quick setup gets you collecting immediately and you can refine details later.</p>
                  <div>
                    <button type="button" onClick={() => void chooseOnboardingPath("quick")}>
                      Start Setup
                    </button>{" "}
                    <button type="button" onClick={() => void chooseOnboardingPath("import")}>
                      Import Existing Collection
                    </button>{" "}
                    <button type="button" onClick={() => void chooseOnboardingPath("sample")}>
                      Use Sample Data
                    </button>
                  </div>
                </div>
              ) : null}
              <div>
                <button type="button" onClick={previousOnboardingStep} disabled={onboardingStep <= 1}>
                  Back Step
                </button>{" "}
                <button
                  type="button"
                  onClick={nextOnboardingStep}
                  disabled={
                    onboardingStep >= ONBOARDING_STEPS.length ||
                    (onboardingStep === 2 && !onboardingIdentityComplete) ||
                    (onboardingStep === 3 && !onboardingStarterDataChoice)
                  }
                >
                  Next Step
                </button>
              </div>
              <p>Complete identity, add your first item, then open the advanced workspace when you are ready.</p>
              <div>
                {onboardingStep === 2 ? (
                  <button type="button" onClick={completeIdentity} disabled={starterIdentityBusy}>
                    {starterIdentityBusy ? "Completing Identity..." : "Complete Identity"}
                  </button>
                ) : null}
                {onboardingStep === 3 ? (
                  <>
                    <button type="button" onClick={() => void chooseOnboardingStarterData("sample")}>
                      Load Sample Data (Recommended)
                    </button>{" "}
                    <button type="button" onClick={() => void chooseOnboardingStarterData("empty")}>
                      Start Empty
                    </button>
                  </>
                ) : null}
                {onboardingCompleted ? (
                  <button type="button" onClick={() => setWorkspaceMode(true)}>
                    Open Advanced Workspace
                  </button>
                ) : null}
              </div>
              {onboardingStep === 4 ? (
                <>
                  <StarterQuickAddForm onSubmit={addStarterItem} isSubmitting={starterSubmitting} />
                </>
              ) : null}
              {onboardingStep === 5 ? (
                <div>
                  <h4>Preferences</h4>
                  <div>
                    <label htmlFor="onboarding-theme">Theme</label>{" "}
                    <select
                      id="onboarding-theme"
                      aria-label="Onboarding theme"
                      value={onboardingTheme}
                      onChange={(e) => setOnboardingTheme((e.target.value as Theme) || "light")}
                    >
                      <option value="light">Light</option>
                      <option value="dark">Dark</option>
                    </select>
                  </div>
                  <div>
                    <label htmlFor="onboarding-backup">Backup frequency</label>{" "}
                    <select
                      id="onboarding-backup"
                      aria-label="Onboarding backup frequency"
                      value={onboardingBackupFrequency}
                      onChange={(e) => setOnboardingBackupFrequency((e.target.value as "daily" | "weekly" | "monthly") || "daily")}
                    >
                      <option value="daily">Daily</option>
                      <option value="weekly">Weekly</option>
                      <option value="monthly">Monthly</option>
                    </select>
                  </div>
                  <div>
                    <label htmlFor="onboarding-scanner">Scanner schedule</label>{" "}
                    <select
                      id="onboarding-scanner"
                      aria-label="Onboarding scanner schedule"
                      value={onboardingScannerPreset}
                      onChange={(e) => setOnboardingScannerPreset((e.target.value as "manual" | "daily" | "weekly") || "manual")}
                    >
                      <option value="manual">Manual only</option>
                      <option value="daily">Daily</option>
                      <option value="weekly">Weekly</option>
                    </select>
                  </div>
                  <button type="button" onClick={finishOnboarding} disabled={onboardingFinishing}>
                    {onboardingFinishing ? "Finishing..." : "Finish Onboarding"}
                  </button>
                </div>
              ) : null}
              <p>Current items: {items.length}</p>
              {itemsError ? <p>Item error: {itemsError}</p> : null}
            </div>
          ) : null}
          {showAdvancedWorkspace && (activeScreen === "all" || activeScreen === "collection") ? (
            <CollectionScreen id="collection">
              <h3>Collection</h3>
              <div>
                <button type="button" onClick={() => setWorkspaceMode(false)}>
                  Back to Starter View
                </button>
              </div>
              <div>
                <input
                  value={collectionQuery.text}
                  onChange={(e) => setCollectionQuery((current) => ({ ...current, text: e.target.value }))}
                  placeholder="Search items"
                  aria-label="Collection search"
                />{" "}
                <input
                  value={collectionQuery.brand}
                  onChange={(e) => setCollectionQuery((current) => ({ ...current, brand: e.target.value }))}
                  placeholder="Brand filter"
                  aria-label="Collection brand filter"
                />{" "}
                <input
                  value={collectionQuery.category}
                  onChange={(e) => setCollectionQuery((current) => ({ ...current, category: e.target.value }))}
                  placeholder="Category filter"
                  aria-label="Collection category filter"
                />{" "}
                <select
                  value={collectionQuery.sort_by}
                  onChange={(e) => setCollectionQuery((current) => ({ ...current, sort_by: e.target.value }))}
                  aria-label="Collection sort"
                >
                  <option value="date_added">Date Added</option>
                  <option value="part_number">Part Number</option>
                  <option value="price">Price</option>
                </select>
              </div>
              <div>
                <input
                  value={savedFilterName}
                  onChange={(e) => setSavedFilterName(e.target.value)}
                  placeholder="Saved filter name"
                  aria-label="Saved filter name"
                />{" "}
                <button type="button" onClick={saveCurrentFilter}>
                  Save Current Filter
                </button>{" "}
                <button type="button" onClick={loadSavedFilters}>
                  Load Saved Filters
                </button>
              </div>
              <ul>
                {savedFilters.map((filter) => (
                  <li key={filter.id}>
                    <button type="button" onClick={() => applySavedFilter(filter)}>
                      {filter.name}
                    </button>
                  </li>
                ))}
              </ul>
              <CollectionItemForm
                key={collectionFormKey}
                onSubmit={addCollectionItem}
                initialValues={collectionFormSeed}
              />
              {itemsLoading ? <p>Loading items...</p> : null}
              {itemsError ? <p>Item error: {itemsError}</p> : null}
              {!itemsLoading && items.length === 0 ? <p>No items found for current filters.</p> : null}
              <table>
                <thead>
                  <tr>
                    <th>Part Number</th>
                    <th>Title</th>
                    <th>Brand</th>
                  </tr>
                </thead>
                <tbody>
                  {items.map((item) => (
                    <tr key={item.id}>
                      <td>
                        <button type="button" onClick={() => setSelectedItemID(item.id)}>
                          {item.part_number}
                        </button>
                      </td>
                      <td>{item.title}</td>
                      <td>{item.brand || "-"}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <h4>Instances</h4>
              <div>
                <button type="button" onClick={loadInstances}>
                  Load Instances
                </button>
              </div>
              <InstanceForm itemID={selectedItemID} onSubmit={addInstance} isSubmitting={instanceSubmitting} />
              <ul>
                {instances.map((instance) => (
                  <li key={instance.id}>
                    {instance.status || "unknown"} / {instance.condition || "unknown"} / qty {String(instance.quantity || 0)}
                  </li>
                ))}
              </ul>
              <h4>Column View</h4>
              <div>
                <div>
                  <p>Brands</p>
                  <ul>
                    {brands.map((brand) => (
                      <li key={brand}>
                        <button
                          type="button"
                          onClick={() => {
                            setColumnBrand(brand);
                            setColumnCategory("");
                            setColumnSeries("");
                          }}
                        >
                          {brand}
                        </button>
                      </li>
                    ))}
                  </ul>
                </div>
                <div>
                  <p>Categories</p>
                  <ul>
                    {categoriesForBrand.map((category) => (
                      <li key={category}>
                        <button
                          type="button"
                          onClick={() => {
                            setColumnCategory(category);
                            setColumnSeries("");
                          }}
                        >
                          {category}
                        </button>
                      </li>
                    ))}
                  </ul>
                </div>
                <div>
                  <p>Series</p>
                  <ul>
                    {seriesForCategory.map((series) => (
                      <li key={series}>
                        <button type="button" onClick={() => setColumnSeries(series)}>
                          {series}
                        </button>
                      </li>
                    ))}
                  </ul>
                </div>
                <div>
                  <p>Items</p>
                  <ul>
                    {columnItems.map((item) => (
                      <li key={item.id}>
                        {item.part_number} - {item.title}
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            </CollectionScreen>
          ) : null}
          {showAdvancedWorkspace && (activeScreen === "all" || activeScreen === "collection") ? (
            <div>
              <h3>Photos</h3>
              <div>
                <input
                  value={selectedItemID}
                  onChange={(e) => setSelectedItemID(e.target.value)}
                  placeholder="Item ID"
                  aria-label="Photo item id"
                />{" "}
                <button type="button" onClick={loadPhotos}>
                  Load Photos
                </button>
              </div>
              <div>
                <input
                  type="file"
                  onChange={(e) => setPhotoFile(e.target.files?.[0] || null)}
                  aria-label="Photo file"
                />{" "}
                <button type="button" onClick={uploadPhoto}>
                  Upload Photo
                </button>
              </div>
              <div
                onDragOver={(e) => e.preventDefault()}
                onDrop={(e) => {
                  e.preventDefault();
                  const dropped = e.dataTransfer.files?.[0];
                  if (dropped) {
                    setPhotoFile(dropped);
                  }
                }}
              >
                Drop photo here to stage upload.
              </div>
              <div>
                <button type="button" onClick={openCamera}>
                  Open Camera
                </button>{" "}
                <button type="button" onClick={() => setCameraOpen(false)}>
                  Close Camera
                </button>
                {cameraOpen ? <p>{cameraStatus}</p> : null}
              </div>
              {photosError ? <p>Photo error: {photosError}</p> : null}
              <ul>
                {photos.map((p) => (
                  <li key={p.id}>
                    {p.filename} {p.is_primary ? "(Primary)" : ""}{" "}
                    <button type="button" onClick={() => setPrimaryPhoto(p.id)}>
                      Set Primary
                    </button>{" "}
                    <button type="button" onClick={() => deletePhoto(p.id)}>
                      Delete
                    </button>{" "}
                    <button type="button" onClick={() => setFullscreenPhoto({ id: p.id, filename: p.filename })}>
                      Open Fullscreen Preview
                    </button>
                  </li>
                ))}
              </ul>
              {fullscreenPhoto ? (
                <div role="dialog" aria-label="Fullscreen photo preview">
                  <p>Fullscreen: {fullscreenPhoto.filename}</p>
                  <button type="button" onClick={() => setFullscreenPhoto(null)}>
                    Close Fullscreen
                  </button>
                </div>
              ) : null}
            </div>
          ) : null}
          {showAdvancedWorkspace && (activeScreen === "all" || activeScreen === "scanner") ? (
            <ScannerScreen>
              <h3>Discovery Scanner</h3>
              <ScannerQuerySetForm
                initialValues={querySetInitialValues}
                onSubmit={createQuerySet}
                isSubmitting={querySetSubmitting}
                onCancel={() => setEditingQuerySetID("")}
              />
              <div>
                <button type="button" onClick={loadQuerySets}>
                  Load Query Sets
                </button>
              </div>
              <div>
                <input
                  value={selectedQuerySetID}
                  onChange={(e) => setSelectedQuerySetID(e.target.value)}
                  placeholder="Query set ID"
                  aria-label="Selected query set id"
                />{" "}
                <button type="button" onClick={runScannerNow}>
                  Run Now
                </button>{" "}
                <button type="button" onClick={runScheduledScans}>
                  Run Scheduled
                </button>{" "}
                <button type="button" onClick={loadCandidates}>
                  Load Candidates
                </button>
              </div>
              <div>
                <button type="button" onClick={loadScannerFailures}>
                  Load Scanner Failures
                </button>{" "}
                <button type="button" onClick={() => loadProviderHealth("ebay")}>
                  Check Provider Health
                </button>{" "}
                <button type="button" onClick={runMatchingNow}>
                  Run Matching
                </button>
              </div>
              {scannerError ? <p>Scanner error: {scannerError}</p> : null}
              {scheduledRunStatus ? <p>Scheduled run: {scheduledRunStatus}</p> : null}
              {scannerRetryStatus ? <p>Scanner retry status: {scannerRetryStatus}</p> : null}
              {matchingRunStatus ? <p>Matching run status: {matchingRunStatus}</p> : null}
              {providerHealth ? <p>Provider health: {providerHealth.provider || "unknown"} / {providerHealth.state || String(providerHealth.healthy)}</p> : null}
              <ul>
                {scannerFailures.map((failure, idx) => (
                  <li key={failure.id || `${failure.query_set_id || "qs"}-${idx}`}>
                    Failure: {failure.query_set_id || "unknown"} / {failure.reason || "n/a"} / attempts {String(failure.attempts || 0)}
                    {failure.query_set_id ? (
                      <>
                        {" "}
                        <button type="button" onClick={() => retryScannerFailure(failure.query_set_id || "")}>
                          Retry Failure {failure.query_set_id}
                        </button>
                      </>
                    ) : null}
                  </li>
                ))}
              </ul>
              <ul>
                {querySets.map((q) => (
                  <li key={q.id}>
                    {q.name} ({q.id}){" "}
                    <button
                      type="button"
                      onClick={() => {
                        setSelectedQuerySetID(q.id);
                        setEditingQuerySetID(q.id);
                      }}
                    >
                      Edit
                    </button>
                  </li>
                ))}
              </ul>
              <ul>
                {candidates.map((c) => (
                  <li key={c.id}>
                    {(c.title || c.listing_id || c.id) + " "}
                    <button type="button" onClick={() => discoveryAction(c.id, "ignore")}>
                      Ignore
                    </button>{" "}
                    <button type="button" onClick={() => discoveryAction(c.id, "add_to_wishlist")}>
                      Wishlist
                    </button>{" "}
                    <button type="button" onClick={() => discoveryAction(c.id, "track_price")}>
                      Track
                    </button>
                  </li>
                ))}
              </ul>
              <div>
                <button type="button" onClick={loadMatchingResults}>
                  Load Matching Results
                </button>
                <p>Matched: {matchedCount}</p>
                <p>Suggested: {suggestedCount}</p>
                <p>Not In Collection: {notInCollectionCount}</p>
              </div>
              <div>
                <input
                  value={notInCollectionFilter.query}
                  onChange={(e) => setNotInCollectionFilter((current) => ({ ...current, query: e.target.value }))}
                  placeholder="Query"
                  aria-label="Not in collection query"
                />{" "}
                <input
                  value={notInCollectionFilter.maxPrice}
                  onChange={(e) => setNotInCollectionFilter((current) => ({ ...current, maxPrice: e.target.value }))}
                  placeholder="Max price"
                  aria-label="Not in collection max price"
                />{" "}
                <input
                  type="date"
                  value={notInCollectionFilter.dateFrom}
                  onChange={(e) => setNotInCollectionFilter((current) => ({ ...current, dateFrom: e.target.value }))}
                  aria-label="Not in collection date from"
                />{" "}
                <button type="button" onClick={loadNotInCollection}>
                  Load Not In Collection
                </button>
              </div>
              <ul>
                {notInCollectionItems.map((item) => (
                  <li key={item.candidate_id}>
                    {item.title || item.candidate_id} - {String(item.price || "")}{" "}
                    <button type="button" onClick={() => notInCollectionAction(item.candidate_id, "ignore")}>
                      Ignore
                    </button>{" "}
                    <button type="button" onClick={() => notInCollectionAction(item.candidate_id, "add_to_wishlist")}>
                      Wishlist
                    </button>{" "}
                    <button type="button" onClick={() => notInCollectionAction(item.candidate_id, "track_price")}>
                      Track
                    </button>{" "}
                    <button type="button" onClick={() => notInCollectionAction(item.candidate_id, "create_item")}>
                      Create Item
                    </button>
                  </li>
                ))}
              </ul>
            </ScannerScreen>
          ) : null}
          {showAdvancedWorkspace && (activeScreen === "all" || activeScreen === "dashboard") ? (
            <DashboardScreen id="dashboard">
              <h3>Dashboard</h3>
              <div>
                <button type="button" onClick={loadDashboard} disabled={dashboardLoading}>
                  {dashboardLoading ? "Refreshing Dashboard..." : "Refresh Dashboard"}
                </button>{" "}
                <button type="button" onClick={() => openWorkspaceFromDashboard("collection")}>
                  Open Collection Workspace
                </button>{" "}
                <button type="button" onClick={() => openWorkspaceFromDashboard("scanner")}>
                  Open Scanner Workspace
                </button>{" "}
                <button type="button" onClick={() => openWorkspaceFromDashboard("pricing")}>
                  Open Pricing Workspace
                </button>{" "}
                <button type="button" onClick={() => openWorkspaceFromDashboard("settings")}>
                  Open Settings Workspace
                </button>
              </div>
              {dashboardLoading ? <p>Loading dashboard...</p> : null}
              {dashboard ? (
                <div>
                  <p>New Discoveries: {String(dashboard.new_discoveries ?? 0)}</p>
                  <p>Wishlist Hits: {String(dashboard.wishlist_hits ?? 0)}</p>
                  <p>Price Drops: {String(dashboard.price_drops ?? 0)}</p>
                  <p>Recently Added: {String(dashboard.recently_added ?? dashboard.recent_items ?? 0)}</p>
                  <p>Total Items: {String(dashboard.total_items ?? 0)}</p>
                  <p>Total Instances: {String(dashboard.total_instances ?? 0)}</p>
                  <p>Estimated Value: {String(dashboard.estimated_value ?? "n/a")}</p>
                </div>
              ) : (
                <p>No dashboard data yet.</p>
              )}
              {insightError ? <p>Insight error: {insightError}</p> : null}
            </DashboardScreen>
          ) : null}
          {showAdvancedWorkspace && (activeScreen === "all" || activeScreen === "pricing") ? (
            <PricingScreen id="pricing">
              <h3>Pricing</h3>
              <div>
                <button type="button" onClick={loadWishlist}>
                  Load Wishlist
                </button>{" "}
                <button type="button" onClick={loadWishlistHits}>
                  Load Wishlist Hits
                </button>{" "}
                <button type="button" onClick={createWishlistEntry}>
                  Add Wishlist Item
                </button>{" "}
                <button type="button" onClick={trackPricingItem}>
                  Track Pricing
                </button>{" "}
                <button type="button" onClick={loadPricingGraph}>
                  Load Pricing Graph
                </button>{" "}
                <button type="button" onClick={loadPricingBySource}>
                  Load Pricing Sources
                </button>{" "}
                <button type="button" onClick={loadPricingHistory}>
                  Load Pricing History
                </button>{" "}
                <button type="button" onClick={loadPricingStats}>
                  Load Pricing Stats
                </button>{" "}
                <button type="button" onClick={loadPricingTrend}>
                  Load Pricing Trend
                </button>{" "}
                <button type="button" onClick={runPricingSnapshot}>
                  Run Pricing Snapshot
                </button>{" "}
                <button type="button" onClick={() => exportText(`/api/pricing/history/export?item_id=${encodeURIComponent(selectedItemID)}`, "pricing_history")}>
                  Export Pricing History
                </button>
              </div>
              <div>
                <input
                  value={wishlistDraft.item_id}
                  onChange={(e) => setWishlistDraft((current) => ({ ...current, item_id: e.target.value }))}
                  placeholder="Wishlist item id"
                  aria-label="Wishlist item id"
                />{" "}
                <input
                  value={wishlistDraft.target_price}
                  onChange={(e) => setWishlistDraft((current) => ({ ...current, target_price: e.target.value }))}
                  placeholder="Target price"
                  aria-label="Wishlist target price"
                />{" "}
                <input
                  value={wishlistDraft.priority}
                  onChange={(e) => setWishlistDraft((current) => ({ ...current, priority: e.target.value }))}
                  placeholder="Priority"
                  aria-label="Wishlist priority"
                />
              </div>
              <ul>
                {wishlist.map((w) => (
                  <li key={w.id}>
                    Wishlist Item: {w.item_id} target {String(w.target_price ?? "")} {w.below_target_now ? "(Below Target)" : ""}{" "}
                    <button type="button" onClick={() => deleteWishlistEntry(w.id)}>
                      Remove
                    </button>
                  </li>
                ))}
              </ul>
              <ul>
                {wishlistHits.map((hit, idx) => (
                  <li key={`${hit.item_id || "item"}-${hit.listing_id || idx}`}>
                    Wishlist Hit: {hit.item_id || "unknown"} / {hit.title || hit.listing_id || "listing"} / {String(hit.price ?? "")}
                  </li>
                ))}
              </ul>
              {pricingTrackStatus ? <p>Pricing track status: {pricingTrackStatus}</p> : null}
              {snapshotStatus ? <p>Snapshot status: {snapshotStatus}</p> : null}
              <p>Pricing points: {pricingPoints.length}</p>
              <p>Pricing history points: {pricingHistory.length}</p>
              <p>Source groups: {Object.keys(pricingBySource).length}</p>
              {pricingStats ? <p>Pricing stats loaded: yes</p> : null}
              {pricingTrend ? <p>Pricing trend loaded: yes</p> : null}
              <ul>
                {Object.entries(pricingBySource).map(([source, snapshots]) => (
                  <li key={source}>
                    {source}: {snapshots.length} snapshots
                  </li>
                ))}
              </ul>
              <p>Export bytes: {exportBytes}</p>
              {insightError ? <p>Insight error: {insightError}</p> : null}
            </PricingScreen>
          ) : null}
          {showAdvancedWorkspace && (activeScreen === "all" || activeScreen === "settings") ? (
            <SettingsScreen id="settings">
              <h3>Settings and Diagnostics</h3>
              <div>
                <h4>Diagnostics</h4>
                <button type="button" onClick={() => toggleDebugMode(true)}>
                  Enable Debug Mode
                </button>{" "}
                <button type="button" onClick={() => toggleDebugMode(false)}>
                  Disable Debug Mode
                </button>{" "}
                <button type="button" onClick={() => exportText("/api/logs/export", "logs")}>
                  Export Logs
                </button>{" "}
                <button type="button" onClick={loadAdminStatus}>
                  Refresh Diagnostics
                </button>
                <button type="button" onClick={loadRuntimeDiagnostics}>
                  Load Runtime Diagnostics
                </button>
                <p>Debug mode: {debugModeEnabled ? "enabled" : "disabled"}</p>
                <p>Runtime channel: {runtimeDiagnostics?.update_channel || "unknown"}</p>
                <p>Runtime signing key configured: {runtimeDiagnostics ? (runtimeDiagnostics.update_public_key_configured ? "yes" : "no") : "unknown"}</p>
                <p>Recovery required: {recoveryDiagnostics ? (recoveryDiagnostics.recovery_required ? "yes" : "no") : "unknown"}</p>
              </div>
              <div>
                <h4>Maintenance</h4>
                <button type="button" onClick={loadProfileSettings}>
                  Load Profile Settings
                </button>{" "}
                <button type="button" onClick={resetIgnoreRules}>
                  Reset Ignore Rules
                </button>{" "}
                <button type="button" onClick={rebuildThumbnails}>
                  Rebuild Thumbnails
                </button>{" "}
                <button type="button" onClick={loadAdminStatus}>
                  Load Admin Status
                </button>{" "}
                <button type="button" onClick={() => runDataMaintenance("/api/data/reindex", "reindex")}>
                  Reindex
                </button>{" "}
                <button type="button" onClick={() => runDataMaintenance("/api/data/repair", "repair")}>
                  Repair
                </button>{" "}
                <button type="button" onClick={() => runDataMaintenance("/api/backup/run", "backup")}>
                  Run Backup
                </button>{" "}
                <button type="button" onClick={loadProfileLicense}>
                  Load Profile License
                </button>{" "}
                <button type="button" onClick={saveProfileLicense}>
                  Save Profile License
                </button>{" "}
                <button type="button" onClick={importLicenseFile}>
                  Import License File
                </button>{" "}
                <button type="button" onClick={() => refreshLicenseStatus()}>
                  Refresh License Status
                </button>{" "}
                <button type="button" onClick={loadBackups}>
                  Load Backups
                </button>{" "}
                <button type="button" onClick={() => exportText("/api/data/export/json", "json")}>
                  Export JSON
                </button>
              </div>
              <div>
                <label htmlFor="profile-license-json">Profile license JSON</label>
                <textarea
                  id="profile-license-json"
                  aria-label="Profile license JSON"
                  value={profileLicenseJSON}
                  onChange={(e) => setProfileLicenseJSON(e.target.value)}
                  rows={3}
                  cols={60}
                />
              </div>
              <div>
                <label htmlFor="license-payload">License payload base64</label>{" "}
                <input
                  id="license-payload"
                  aria-label="License payload base64"
                  value={licenseImportDraft.payload_base64}
                  onChange={(e) => setLicenseImportDraft((current) => ({ ...current, payload_base64: e.target.value }))}
                />{" "}
                <label htmlFor="license-signature">License signature base64</label>{" "}
                <input
                  id="license-signature"
                  aria-label="License signature base64"
                  value={licenseImportDraft.signature_base64}
                  onChange={(e) => setLicenseImportDraft((current) => ({ ...current, signature_base64: e.target.value }))}
                />
              </div>
              <div>
                <label htmlFor="backup-selection">Available backups</label>{" "}
                <select
                  id="backup-selection"
                  aria-label="Backup selection"
                  value={selectedBackupPath}
                  onChange={(e) => setSelectedBackupPath(e.target.value)}
                >
                  <option value="">Select backup</option>
                  {backupEntries.map((entry) => (
                    <option key={entry.path} value={entry.path}>
                      {entry.name}
                    </option>
                  ))}
                </select>{" "}
                <label>
                  <input type="checkbox" checked={confirmRestore} onChange={(e) => setConfirmRestore(e.target.checked)} aria-label="Confirm restore" /> Confirm
                  restore
                </label>{" "}
                <button type="button" onClick={restoreSelectedBackup}>
                  Restore Selected Backup
                </button>
                <p>Backup count: {backupEntries.length}</p>
                <ul>
                  {backupEntries.map((entry) => (
                    <li key={entry.path}>
                      {entry.name} ({entry.timestampLabel})
                    </li>
                  ))}
                </ul>
              </div>
              <div>
                <ProfileSettingsForm initialValues={profileSettingsInitial} onSubmit={saveProfileSettings} />
                <SecretsForm initialValues={secretsInitial} onSubmit={saveSecrets} />
                <DataImportExportWizard onDryRun={dataImportDryRun} onApply={dataImportApply} onExport={dataExportRun} />
              </div>
              {licenseStatus ? <p>License: {licenseStatus.state || "unknown"} / {licenseStatus.tier || "unknown"}</p> : null}
              {licenseStatus ? <p>License validation: {licenseStatus.state || "unknown"} / {licenseStatus.tier || "unknown"}</p> : null}
              {licenseStatus?.features?.length ? <p>License features: {licenseStatus.features.join(", ")}</p> : null}
              {licenseStatus?.expires_at ? <p>License expires: {licenseStatus.expires_at}</p> : null}
              {licenseImportStatus ? <p>License import status: {licenseImportStatus}</p> : null}
              <p>Log entries: {logCount}</p>
              <ul>
                {activityLogs.map((entry, idx) => (
                  <li key={`${String(entry.event || "event")}-${idx}`}>
                    Activity: {String(entry.event || "unknown")} {entry.created_at ? `(${String(entry.created_at)})` : ""}
                  </li>
                ))}
              </ul>
              <p>Settings status: {settingsStatus || "idle"}</p>
              {adminError ? <p>Admin error: {adminError}</p> : null}
              {adminError === "failed_to_restore_backup" ? <p>Restore failed: verify the selected backup file is valid and readable.</p> : null}
            </SettingsScreen>
          ) : null}
          {showAdvancedWorkspace && (activeScreen === "all" || activeScreen === "collection") ? (
            <div>
              <h3>Barcodes</h3>
              <div>
                <input
                  value={barcodeInput}
                  onChange={(e) => setBarcodeInput(e.target.value)}
                  placeholder="Barcode"
                  aria-label="Barcode"
                />{" "}
                <button type="button" onClick={addBarcode}>
                  Add Barcode
                </button>{" "}
                <button type="button" onClick={loadItemBarcodes}>
                  Load Barcodes
                </button>{" "}
                <button type="button" onClick={lookupBarcode}>
                  Lookup Barcode
                </button>{" "}
                <button type="button" onClick={externalBarcodeSearch}>
                  External Search
                </button>
              </div>
              <p>Detect from image: use photos upload + lookup flow (UI hook ready).</p>
              {barcodeError ? <p>Barcode error: {barcodeError}</p> : null}
              <ul>
                {barcodes.map((b, idx) => (
                  <li key={b.id || `${b.barcode}-${idx}`}>{b.barcode}</li>
                ))}
              </ul>
              <p>Local matches: {barcodeLookupMatches.length}</p>
              {barcodeExternalURL ? <p>{barcodeExternalURL}</p> : null}
            </div>
          ) : null}
          {showAdvancedWorkspace && (activeScreen === "all" || activeScreen === "collection") ? (
            <div>
              <h3>AI Assist</h3>
              <AIAssistForms
                aiEnabled={aiEnabled}
                aiError={aiError}
                suggestion={aiSuggestion}
                lastAction={aiLastAction}
                onToggle={toggleAI}
                onTest={testAI}
                onSuggestTitle={suggestFromTitle}
                onSuggestPhoto={suggestFromPhoto}
                onApplySuggestion={applySuggestion}
                onRetry={retryLastAIAction}
              />
            </div>
          ) : null}
          {error ? <p>Profile error: {error}</p> : null}
          <ul>
            <li>
              <a href="/healthz">Health Check</a>
            </li>
            <li>
              <a href="/api/runtime">Runtime</a>
            </li>
            <li>
              <a href="/api/runtime/recovery">Recovery State</a>
            </li>
              <li>
                <a href="/apidocs">API Kitchen Sync</a>
              </li>
          </ul>
        </section>
      </section>
    </main>
  );
}
