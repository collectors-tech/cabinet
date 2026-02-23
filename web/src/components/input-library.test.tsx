import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import {
  CheckboxInput,
  CurrencyInput,
  DateInput,
  MultiSelectInput,
  SelectInput,
  SwitchInput,
  TextAreaInput,
  TextInput,
} from "./forms";

describe("input component library", () => {
  it("renders all core form controls", () => {
    render(
      <div>
        <TextInput aria-label="TextInput" defaultValue="abc" />
        <TextAreaInput aria-label="TextAreaInput" defaultValue="notes" />
        <SelectInput aria-label="SelectInput" defaultValue="one">
          <option value="one">One</option>
          <option value="two">Two</option>
        </SelectInput>
        <MultiSelectInput
          aria-label="MultiSelectInput"
          options={[
            { label: "One", value: "one" },
            { label: "Two", value: "two" },
          ]}
          value={["one"]}
          onChange={() => undefined}
        />
        <CheckboxInput aria-label="CheckboxInput" checked onChange={() => undefined} />
        <SwitchInput aria-label="SwitchInput" checked onChange={() => undefined} />
        <DateInput aria-label="DateInput" defaultValue="2026-02-23" />
        <CurrencyInput aria-label="CurrencyInput" defaultValue="12.30" />
      </div>,
    );

    expect(screen.getByLabelText("TextInput")).toBeInTheDocument();
    expect(screen.getByLabelText("TextAreaInput")).toBeInTheDocument();
    expect(screen.getByLabelText("SelectInput")).toBeInTheDocument();
    expect(screen.getByLabelText("MultiSelectInput")).toBeInTheDocument();
    expect(screen.getByLabelText("CheckboxInput")).toBeInTheDocument();
    expect(screen.getByLabelText("SwitchInput")).toBeInTheDocument();
    expect(screen.getByLabelText("DateInput")).toBeInTheDocument();
    expect(screen.getByLabelText("CurrencyInput")).toBeInTheDocument();
  });

  it("supports disabled and error states", () => {
    render(
      <div>
        <TextInput aria-label="disabled-input" disabled />
        <TextAreaInput aria-label="error-area" data-invalid="true" />
      </div>,
    );
    expect(screen.getByLabelText("disabled-input")).toBeDisabled();
    expect(screen.getByLabelText("error-area")).toHaveAttribute("data-invalid", "true");
  });

  it("fires interactions for switch and multi-select", () => {
    let switched = false;
    let selected: string[] = [];
    render(
      <div>
        <SwitchInput
          aria-label="toggle"
          checked={false}
          onChange={(next) => {
            switched = next;
          }}
        />
        <MultiSelectInput
          aria-label="multi"
          options={[
            { label: "One", value: "one" },
            { label: "Two", value: "two" },
          ]}
          value={[]}
          onChange={(values) => {
            selected = values;
          }}
        />
      </div>,
    );

    fireEvent.click(screen.getByLabelText("toggle"));
    const multi = screen.getByLabelText("multi") as HTMLSelectElement;
    multi.options[1].selected = true;
    fireEvent.change(multi);

    expect(switched).toBe(true);
    expect(selected).toEqual(["two"]);
  });
});
