import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { BoundTextInput, Form, SubmitButton } from "./forms";

const starterQuickAddSchema = z.object({
  part_number: z.string().trim().min(1, "Part number is required"),
  title: z.string().trim().min(1, "Title is required"),
  brand: z.string().trim().optional(),
});

export type StarterQuickAddValues = z.infer<typeof starterQuickAddSchema>;

export function StarterQuickAddForm({
  onSubmit,
  isSubmitting,
}: {
  onSubmit: (values: StarterQuickAddValues) => Promise<void> | void;
  isSubmitting?: boolean;
}) {
  const form = useForm<StarterQuickAddValues>({
    resolver: zodResolver(starterQuickAddSchema),
    defaultValues: {
      part_number: "",
      title: "",
      brand: "",
    },
  });

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(async (values) => {
          await onSubmit(values);
          form.reset({ part_number: "", title: "", brand: "" });
        })}
      >
        <BoundTextInput<StarterQuickAddValues, "part_number">
          name="part_number"
          label="Part Number"
          placeholder="Part Number"
        />
        <BoundTextInput<StarterQuickAddValues, "title">
          name="title"
          label="Item Title"
          placeholder="Item Title"
        />
        <BoundTextInput<StarterQuickAddValues, "brand">
          name="brand"
          label="Brand"
          placeholder="Brand"
        />
        <SubmitButton isSubmitting={Boolean(isSubmitting)}>Add First Item</SubmitButton>
      </form>
    </Form>
  );
}

