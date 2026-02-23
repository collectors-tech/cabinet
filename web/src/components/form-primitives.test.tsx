import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  SubmitButton,
  TextInput,
} from "./forms";

const schema = z.object({
  part_number: z.string().min(3, "Part number is required"),
});

function TestForm() {
  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: { part_number: "" },
    mode: "onSubmit",
  });

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(() => undefined)}>
        <FormField
          control={form.control}
          name="part_number"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Part Number</FormLabel>
              <FormControl>
                <TextInput aria-label="Part Number" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <SubmitButton isSubmitting={false}>Save</SubmitButton>
      </form>
    </Form>
  );
}

describe("form primitives", () => {
  it("renders accessible field and submit button", () => {
    render(<TestForm />);
    expect(screen.getByLabelText(/part number/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /save/i })).toBeInTheDocument();
  });
});

