import { useEffect, useState } from "react";

type Theme = "light" | "dark";

function detectInitialTheme(): Theme {
  const saved = localStorage.getItem("cabinet.theme");
  if (saved === "dark" || saved === "light") {
    return saved;
  }
  return "light";
}

export function App() {
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
  const [itemsError, setItemsError] = useState("");
  const [newItem, setNewItem] = useState({
    part_number: "",
    title: "",
    brand: "",
    category: "General",
  });
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
  const [photosError, setPhotosError] = useState("");
  const [querySets, setQuerySets] = useState<Array<{ id: string; name: string }>>([]);
  const [selectedQuerySetID, setSelectedQuerySetID] = useState("");
  const [candidates, setCandidates] = useState<Array<{ id: string; title?: string; listing_id?: string; status?: string }>>([]);
  const [matchingResults, setMatchingResults] = useState<Array<{ candidate_id?: string; state?: string; part_number?: string; item_id?: string }>>(
    [],
  );
  const [notInCollectionItems, setNotInCollectionItems] = useState<Array<{ candidate_id: string; title?: string; price?: number; url?: string; last_seen?: string }>>(
    [],
  );
  const [notInCollectionFilter, setNotInCollectionFilter] = useState({ query: "", maxPrice: "", dateFrom: "" });
  const [scannerError, setScannerError] = useState("");
  const [newQuerySet, setNewQuerySet] = useState({ name: "", keywords: "afx" });
  const [dashboard, setDashboard] = useState<Record<string, unknown> | null>(null);
  const [wishlist, setWishlist] = useState<Array<{ id: string; item_id: string; target_price?: number }>>([]);
  const [pricingPoints, setPricingPoints] = useState<Array<{ day?: string; date?: string; price?: number; min?: number; median?: number; latest?: number }>>([]);
  const [insightError, setInsightError] = useState("");
  const [licenseStatus, setLicenseStatus] = useState<{ state?: string; tier?: string } | null>(null);
  const [logCount, setLogCount] = useState(0);
  const [adminError, setAdminError] = useState("");
  const [settingsStatus, setSettingsStatus] = useState("");
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
  const [authStatus, setAuthStatus] = useState("");
  const [authSessionID, setAuthSessionID] = useState("");
  const [requiresRegistration, setRequiresRegistration] = useState<boolean | null>(null);
  const [credentialJSON, setCredentialJSON] = useState("{}");
  const [sessionToken, setSessionToken] = useState("");
  const [recoveryPassphrase, setRecoveryPassphrase] = useState("");

  useEffect(() => {
    document.body.setAttribute("data-theme", theme);
    localStorage.setItem("cabinet.theme", theme);
  }, [theme]);

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
      await loadItems();
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_activate_profile");
    }
  }

  async function addItem() {
    setItemsError("");
    if (!newItem.part_number.trim() || !newItem.title.trim()) {
      setItemsError("part_number_and_title_required");
      return;
    }
    try {
      const resp = await fetch("/api/items", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(newItem),
      });
      if (!resp.ok) {
        throw new Error("failed_to_create_item");
      }
      const created = (await resp.json()) as { id: string; part_number: string; title: string; brand?: string; category?: string; series?: string };
      setItems((current) => [...current, created]);
      if (!selectedItemID) {
        setSelectedItemID(created.id);
      }
      setNewItem({ part_number: "", title: "", brand: "", category: "General" });
    } catch (e) {
      setItemsError(e instanceof Error ? e.message : "failed_to_create_item");
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
      const data = (await resp.json()) as { query_sets?: Array<{ id: string; name: string }> };
      const listed = data.query_sets || [];
      setQuerySets(listed);
      if (listed.length > 0 && !selectedQuerySetID) {
        setSelectedQuerySetID(listed[0].id);
      }
    } catch (e) {
      setScannerError(e instanceof Error ? e.message : "failed_to_list_query_sets");
    }
  }

  async function createQuerySet() {
    setScannerError("");
    try {
      const payload = {
        name: newQuerySet.name || "Query Set",
        keywords: newQuerySet.keywords
          .split(",")
          .map((k) => k.trim())
          .filter(Boolean),
      };
      const resp = await fetch("/api/scanner/query-sets", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      if (!resp.ok) {
        throw new Error("failed_to_create_query_set");
      }
      const created = (await resp.json()) as { id: string; name: string };
      setQuerySets((current) => [...current, created]);
      if (!selectedQuerySetID) {
        setSelectedQuerySetID(created.id);
      }
      setNewQuerySet({ name: "", keywords: "afx" });
    } catch (e) {
      setScannerError(e instanceof Error ? e.message : "failed_to_create_query_set");
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
    try {
      const resp = await fetch("/api/dashboard");
      if (!resp.ok) {
        throw new Error("failed_to_load_dashboard");
      }
      const data = (await resp.json()) as Record<string, unknown>;
      setDashboard(data);
    } catch (e) {
      setInsightError(e instanceof Error ? e.message : "failed_to_load_dashboard");
    }
  }

  async function loadWishlist() {
    setInsightError("");
    try {
      const resp = await fetch("/api/wishlist");
      if (!resp.ok) {
        throw new Error("failed_to_load_wishlist");
      }
      const data = (await resp.json()) as { wishlist?: Array<{ id: string; item_id: string; target_price?: number }> };
      setWishlist(data.wishlist || []);
    } catch (e) {
      setInsightError(e instanceof Error ? e.message : "failed_to_load_wishlist");
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

  async function loadAdminStatus() {
    if (!activeProfile?.id) {
      return;
    }
    setAdminError("");
    try {
      const licenseResp = await fetch(`/api/license/status?profile_id=${encodeURIComponent(activeProfile.id)}`);
      if (!licenseResp.ok) {
        throw new Error("failed_to_load_license_status");
      }
      const license = (await licenseResp.json()) as { state?: string; tier?: string };
      setLicenseStatus(license);

      const logsResp = await fetch("/api/logs/activity?limit=10");
      if (!logsResp.ok) {
        throw new Error("failed_to_load_activity_logs");
      }
      const logs = (await logsResp.json()) as { activity?: Array<unknown> };
      setLogCount((logs.activity || []).length);
    } catch (e) {
      setAdminError(e instanceof Error ? e.message : "failed_to_load_admin_status");
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

  async function suggestFromTitle() {
    if (!activeProfile?.id || !aiTitleInput.trim()) {
      setAiError("profile_and_title_required");
      return;
    }
    setAiError("");
    try {
      const resp = await fetch("/api/ai/suggest/title", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ profile_id: activeProfile.id, title: aiTitleInput.trim() }),
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

  async function suggestFromPhoto() {
    if (!activeProfile?.id || !aiPhotoURL.trim()) {
      setAiError("profile_and_photo_required");
      return;
    }
    setAiError("");
    try {
      const resp = await fetch("/api/ai/suggest/photo", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ profile_id: activeProfile.id, image_url: aiPhotoURL.trim() }),
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
    if (!window.confirm("Apply AI suggestion to item title draft?")) {
      return;
    }
    setNewItem((current) => ({ ...current, title: String(aiSuggestion.title) }));
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
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_finish_login");
    }
  }

  async function saveRecoveryPassphrase() {
    if (!activeProfile?.id || !recoveryPassphrase) {
      return;
    }
    setError("");
    try {
      await postJSON("/api/auth/recovery/passphrase", { profile_id: activeProfile.id, passphrase: recoveryPassphrase });
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

  async function validateSession() {
    if (!sessionToken) {
      return;
    }
    setError("");
    try {
      await postJSON("/api/auth/session/validate", { session_token: sessionToken });
      setAuthStatus("session_valid");
    } catch (e) {
      setError(e instanceof Error ? e.message : "session_invalid_or_locked");
    }
  }

  async function lockSession() {
    if (!sessionToken) {
      return;
    }
    setError("");
    try {
      await postJSON("/api/auth/session/lock", { session_token: sessionToken });
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

  return (
    <main data-testid="app-shell" className="cabinet-shell">
      <aside className="cabinet-sidebar">
        <h1>Cabinet</h1>
        <nav>
          <a href="#dashboard">Dashboard</a>
          <a href="#collection">Collection</a>
          <a href="#scanner">Scanner</a>
          <a href="#pricing">Pricing</a>
          <a href="#settings">Settings</a>
        </nav>
      </aside>
      <section className="cabinet-content">
        <header className="cabinet-topbar">
          <strong>Runtime connected. UI foundation active.</strong>
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
          <p>Next: onboarding, WebAuthn flows, collection workspace, and Cypress E2E.</p>
          {loading ? <p>Loading profiles...</p> : null}
          {!loading && profiles.length === 0 ? (
            <div>
              <p>No local profiles found. Create your first profile to continue.</p>
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
              <div>
                <input
                  value={recoveryPassphrase}
                  onChange={(e) => setRecoveryPassphrase(e.target.value)}
                  placeholder="Recovery passphrase"
                  aria-label="Recovery passphrase"
                />{" "}
                <button type="button" onClick={saveRecoveryPassphrase}>
                  Save Recovery Passphrase
                </button>{" "}
                <button type="button" onClick={beginRecoveryReset}>
                  Begin Recovery Reset
                </button>
              </div>
              <div>
                <input
                  value={sessionToken}
                  onChange={(e) => setSessionToken(e.target.value)}
                  placeholder="Session token"
                  aria-label="Session token"
                />{" "}
                <button type="button" onClick={validateSession}>
                  Validate Session
                </button>{" "}
                <button type="button" onClick={lockSession}>
                  Lock Session
                </button>
              </div>
              <p>Auth status: {authStatus || "idle"}</p>
            </div>
          ) : null}
          {activeProfile ? (
            <div>
              <h3>Collection</h3>
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
              <div>
                <input
                  value={newItem.part_number}
                  onChange={(e) => setNewItem((current) => ({ ...current, part_number: e.target.value }))}
                  placeholder="Part Number"
                  aria-label="Part number"
                />{" "}
                <input
                  value={newItem.title}
                  onChange={(e) => setNewItem((current) => ({ ...current, title: e.target.value }))}
                  placeholder="Item Title"
                  aria-label="Item title"
                />{" "}
                <input
                  value={newItem.brand}
                  onChange={(e) => setNewItem((current) => ({ ...current, brand: e.target.value }))}
                  placeholder="Brand"
                  aria-label="Brand"
                />{" "}
                <input
                  value={newItem.category}
                  onChange={(e) => setNewItem((current) => ({ ...current, category: e.target.value }))}
                  placeholder="Category"
                  aria-label="Category"
                />{" "}
                <button type="button" onClick={addItem}>
                  Add Item
                </button>
              </div>
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
                      <td>{item.part_number}</td>
                      <td>{item.title}</td>
                      <td>{item.brand || "-"}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
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
            </div>
          ) : null}
          {activeProfile ? (
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
                    </button>
                  </li>
                ))}
              </ul>
            </div>
          ) : null}
          {activeProfile ? (
            <div>
              <h3>Discovery Scanner</h3>
              <div>
                <input
                  value={newQuerySet.name}
                  onChange={(e) => setNewQuerySet((current) => ({ ...current, name: e.target.value }))}
                  placeholder="Query set name"
                  aria-label="Query set name"
                />{" "}
                <input
                  value={newQuerySet.keywords}
                  onChange={(e) => setNewQuerySet((current) => ({ ...current, keywords: e.target.value }))}
                  placeholder="Keywords comma separated"
                  aria-label="Query set keywords"
                />{" "}
                <button type="button" onClick={createQuerySet}>
                  Create Query Set
                </button>{" "}
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
                <button type="button" onClick={loadCandidates}>
                  Load Candidates
                </button>
              </div>
              {scannerError ? <p>Scanner error: {scannerError}</p> : null}
              <ul>
                {querySets.map((q) => (
                  <li key={q.id}>
                    {q.name} ({q.id})
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
            </div>
          ) : null}
          {activeProfile ? (
            <div>
              <h3>Dashboard and Pricing</h3>
              <div>
                <button type="button" onClick={loadDashboard}>
                  Load Dashboard
                </button>{" "}
                <button type="button" onClick={loadWishlist}>
                  Load Wishlist
                </button>{" "}
                <button type="button" onClick={loadPricingGraph}>
                  Load Pricing Graph
                </button>{" "}
                <button type="button" onClick={() => exportText(`/api/pricing/history/export?item_id=${encodeURIComponent(selectedItemID)}`, "pricing_history")}>
                  Export Pricing History
                </button>
              </div>
              {dashboard ? (
                <div>
                  <p>New Discoveries: {String(dashboard.new_discoveries ?? 0)}</p>
                  <p>Wishlist Hits: {String(dashboard.wishlist_hits ?? 0)}</p>
                  <p>Price Drops: {String(dashboard.price_drops ?? 0)}</p>
                  <p>Total Items: {String(dashboard.total_items ?? 0)}</p>
                </div>
              ) : null}
              <ul>
                {wishlist.map((w) => (
                  <li key={w.id}>Wishlist Item: {w.item_id} target {String(w.target_price ?? "")}</li>
                ))}
              </ul>
              <p>Pricing points: {pricingPoints.length}</p>
              <p>Export bytes: {exportBytes}</p>
              {insightError ? <p>Insight error: {insightError}</p> : null}
            </div>
          ) : null}
          {activeProfile ? (
            <div>
              <h3>Settings and Diagnostics</h3>
              <div>
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
                <button type="button" onClick={() => exportText("/api/logs/export", "logs")}>
                  Export Logs
                </button>{" "}
                <button type="button" onClick={() => exportText("/api/data/export/json", "json")}>
                  Export JSON
                </button>
              </div>
              {licenseStatus ? <p>License: {licenseStatus.state || "unknown"} / {licenseStatus.tier || "unknown"}</p> : null}
              <p>Log entries: {logCount}</p>
              <p>Settings status: {settingsStatus || "idle"}</p>
              {adminError ? <p>Admin error: {adminError}</p> : null}
            </div>
          ) : null}
          {activeProfile ? (
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
          {activeProfile ? (
            <div>
              <h3>AI Assist</h3>
              <div>
                <button type="button" onClick={() => toggleAI(true)}>
                  Enable AI
                </button>{" "}
                <button type="button" onClick={() => toggleAI(false)}>
                  Disable AI
                </button>{" "}
                <button type="button" onClick={testAI}>
                  Test AI
                </button>
              </div>
              <p>AI enabled: {String(aiEnabled)}</p>
              <div>
                <input
                  value={aiTitleInput}
                  onChange={(e) => setAiTitleInput(e.target.value)}
                  placeholder="Listing title"
                  aria-label="AI title input"
                />{" "}
                <button type="button" onClick={suggestFromTitle}>
                  Suggest From Title
                </button>
              </div>
              <div>
                <input
                  value={aiPhotoURL}
                  onChange={(e) => setAiPhotoURL(e.target.value)}
                  placeholder="Image URL"
                  aria-label="AI photo url"
                />{" "}
                <button type="button" onClick={suggestFromPhoto}>
                  Suggest From Photo
                </button>
              </div>
              {aiSuggestion ? (
                <div>
                  <p>AI confidence: {String(aiSuggestion.confidence ?? "")}</p>
                  <p>AI title: {String(aiSuggestion.title ?? "")}</p>
                  <button type="button" onClick={applySuggestion}>
                    Apply Suggestion
                  </button>
                </div>
              ) : null}
              {aiError ? <p>AI error: {aiError}</p> : null}
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
          </ul>
        </section>
      </section>
    </main>
  );
}
