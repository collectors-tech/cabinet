import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { BoundTextInput, Form, SelectInput, SubmitButton } from "./forms";

const profileSettingsSchema = z.object({
  scanner_schedule: z.string().trim().min(1, "Scanner schedule is required"),
  backup_frequency: z.enum(["hourly", "daily", "weekly", "manual"]),
  db_path: z.string().trim().min(1, "Database path is required"),
  update_channel: z.enum(["stable", "beta", "canary"]).default("stable"),
});

const secretsSchema = z.object({
  openai_api_key: z.string().trim().min(1, "OpenAI API key is required"),
  ebay_app_id: z.string().trim().min(1, "eBay app id is required"),
  ebay_auth_token: z.string().trim().min(1, "eBay auth token is required"),
});

export type ProfileSettingsValues = z.infer<typeof profileSettingsSchema>;
export type SecretsValues = z.infer<typeof secretsSchema>;

export function ProfileSettingsForm({
  initialValues,
  onSubmit,
  isSubmitting,
}: {
  initialValues?: Partial<ProfileSettingsValues>;
  onSubmit: (values: ProfileSettingsValues) => Promise<void> | void;
  isSubmitting?: boolean;
}) {
  const form = useForm<ProfileSettingsValues>({
    resolver: zodResolver(profileSettingsSchema),
    defaultValues: {
      scanner_schedule: initialValues?.scanner_schedule || "",
      backup_frequency: initialValues?.backup_frequency || "daily",
      db_path: initialValues?.db_path || "",
      update_channel: initialValues?.update_channel || "stable",
    },
  });

  useEffect(() => {
    form.reset({
      scanner_schedule: initialValues?.scanner_schedule || "",
      backup_frequency: initialValues?.backup_frequency || "daily",
      db_path: initialValues?.db_path || "",
      update_channel: initialValues?.update_channel || "stable",
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [initialValues?.scanner_schedule, initialValues?.backup_frequency, initialValues?.db_path, initialValues?.update_channel]);

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(async (values) => {
          await onSubmit(values);
        })}
      >
        <BoundTextInput<ProfileSettingsValues, "scanner_schedule">
          name="scanner_schedule"
          label="Scanner Schedule"
          placeholder="0 6 * * *"
        />
        <div className="cabinet-form-item">
          <label className="cabinet-form-label" htmlFor="backup-frequency">
            Backup Frequency
          </label>
          <SelectInput id="backup-frequency" aria-label="Backup frequency" {...form.register("backup_frequency")}>
            <option value="hourly">hourly</option>
            <option value="daily">daily</option>
            <option value="weekly">weekly</option>
            <option value="manual">manual</option>
          </SelectInput>
        </div>
        <BoundTextInput<ProfileSettingsValues, "db_path">
          name="db_path"
          label="Database Path"
          placeholder="C:/Cabinet/profiles/default/cabinet.db"
        />
        <div className="cabinet-form-item">
          <label className="cabinet-form-label" htmlFor="update-channel">
            Update Channel
          </label>
          <SelectInput id="update-channel" aria-label="Update channel" {...form.register("update_channel")}>
            <option value="stable">stable</option>
            <option value="beta">beta</option>
            <option value="canary">canary</option>
          </SelectInput>
        </div>
        <SubmitButton isSubmitting={Boolean(isSubmitting)}>Save Profile Settings</SubmitButton>
      </form>
    </Form>
  );
}

export function SecretsForm({
  onSubmit,
  initialValues,
  isSubmitting,
}: {
  onSubmit: (values: SecretsValues) => Promise<void> | void;
  initialValues?: Partial<SecretsValues>;
  isSubmitting?: boolean;
}) {
  const [reveal, setReveal] = useState(false);
  const form = useForm<SecretsValues>({
    resolver: zodResolver(secretsSchema),
    defaultValues: {
      openai_api_key: initialValues?.openai_api_key || "",
      ebay_app_id: initialValues?.ebay_app_id || "",
      ebay_auth_token: initialValues?.ebay_auth_token || "",
    },
  });

  useEffect(() => {
    form.reset({
      openai_api_key: initialValues?.openai_api_key || "",
      ebay_app_id: initialValues?.ebay_app_id || "",
      ebay_auth_token: initialValues?.ebay_auth_token || "",
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [initialValues?.openai_api_key, initialValues?.ebay_app_id, initialValues?.ebay_auth_token]);

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(async (values) => {
          await onSubmit(values);
        })}
      >
        <BoundTextInput<SecretsValues, "openai_api_key">
          name="openai_api_key"
          label="OpenAI API Key"
          placeholder="sk-..."
          type={reveal ? "text" : "password"}
        />
        <BoundTextInput<SecretsValues, "ebay_app_id">
          name="ebay_app_id"
          label="eBay App ID"
          placeholder="eBay app id"
          type={reveal ? "text" : "password"}
        />
        <BoundTextInput<SecretsValues, "ebay_auth_token">
          name="ebay_auth_token"
          label="eBay Auth Token"
          placeholder="eBay auth token"
          type={reveal ? "text" : "password"}
        />
        <button type="button" onClick={() => setReveal((current) => !current)}>
          {reveal ? "Hide Secrets" : "Show Secrets"}
        </button>
        <SubmitButton isSubmitting={Boolean(isSubmitting)}>Save Secrets</SubmitButton>
      </form>
    </Form>
  );
}
