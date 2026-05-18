/**
 * AuraSelector - Searchable dropdown for selecting multiple auras to display
 */

import { useState, useRef, useEffect, useMemo } from "react";
import { ChevronDown, Search, X, Check } from "lucide-react";
import { cn } from "@/lib/utils";

interface AuraSelectorProps {
  auras: string[];
  /** Comma-separated list of selected auras */
  selected: string | null;
  /** Called with comma-separated list or null */
  onChange: (auras: string | null) => void;
  className?: string;
}

/** Parse comma-separated aura string to Set */
function parseSelectedAuras(selected: string | null): Set<string> {
  if (!selected) return new Set();
  return new Set(selected.split(",").filter(Boolean));
}

/** Serialize Set of auras to comma-separated string */
function serializeSelectedAuras(auras: Set<string>): string | null {
  if (auras.size === 0) return null;
  return [...auras].sort().join(",");
}

/**
 * Simple fuzzy match - checks if all characters in pattern appear in str in order
 */
function fuzzyMatch(pattern: string, str: string): { match: boolean; score: number } {
  const patternLower = pattern.toLowerCase();
  const strLower = str.toLowerCase();

  if (patternLower.length === 0) return { match: true, score: 0 };

  let patternIdx = 0;
  let score = 0;
  let consecutiveBonus = 0;

  for (let i = 0; i < strLower.length && patternIdx < patternLower.length; i++) {
    if (strLower[i] === patternLower[patternIdx]) {
      score += 1 + consecutiveBonus;
      consecutiveBonus += 1;
      if (i === 0 || str[i - 1] === " ") {
        score += 2;
      }
      patternIdx++;
    } else {
      consecutiveBonus = 0;
    }
  }

  return {
    match: patternIdx === patternLower.length,
    score,
  };
}

export function AuraSelector({ auras, selected, onChange, className }: AuraSelectorProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const containerRef = useRef<HTMLDivElement>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);
  
  // Parse selected auras into a Set for easy lookup
  const selectedSet = useMemo(() => parseSelectedAuras(selected), [selected]);

  // Close on outside click
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
        setSearchQuery("");
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Focus search input when opened
  useEffect(() => {
    if (isOpen && searchInputRef.current) {
      searchInputRef.current.focus();
    }
  }, [isOpen]);

  // Filter and sort auras based on search (selected items first when not searching)
  const filteredAuras = useMemo(() => {
    let results: { aura: string; score: number; isSelected: boolean }[];
    
    if (!searchQuery.trim()) {
      // When no search, show selected first, then alphabetically
      results = auras.map(aura => ({
        aura,
        score: 0,
        isSelected: selectedSet.has(aura),
      }));
      results.sort((a, b) => {
        if (a.isSelected !== b.isSelected) return a.isSelected ? -1 : 1;
        return a.aura.localeCompare(b.aura);
      });
    } else {
      results = [];
      for (const aura of auras) {
        const { match, score } = fuzzyMatch(searchQuery, aura);
        if (match) {
          results.push({ aura, score, isSelected: selectedSet.has(aura) });
        }
      }
      // Sort by score descending, then alphabetically
      results.sort((a, b) => b.score - a.score || a.aura.localeCompare(b.aura));
    }

    return results.map(r => r.aura);
  }, [auras, searchQuery, selectedSet]);

  const handleToggle = (aura: string) => {
    const newSet = new Set(selectedSet);
    if (newSet.has(aura)) {
      newSet.delete(aura);
    } else {
      newSet.add(aura);
    }
    onChange(serializeSelectedAuras(newSet));
    // Don't close dropdown - allow multiple selections
  };

  const handleClear = (e: React.MouseEvent) => {
    e.stopPropagation();
    onChange(null);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Escape") {
      setIsOpen(false);
      setSearchQuery("");
    }
  };

  // Display text for trigger button
  const displayText = useMemo(() => {
    if (selectedSet.size === 0) return null;
    if (selectedSet.size === 1) return [...selectedSet][0];
    return `${selectedSet.size} auras selected`;
  }, [selectedSet]);

  return (
    <div ref={containerRef} className={cn("relative", className)}>
      {/* Trigger button */}
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          "flex items-center gap-1.5 text-xs px-2 py-1 rounded border",
          "bg-transparent cursor-pointer hover:bg-muted/50 transition-colors",
          "min-w-[140px] max-w-[240px]"
        )}
      >
        <Search className="h-3 w-3 text-muted-foreground shrink-0" />
        <span className={cn("flex-1 text-left truncate", !displayText && "text-muted-foreground")}>
          {displayText ?? "Select auras..."}
        </span>
        {selectedSet.size > 0 ? (
          <X
            className="h-3 w-3 text-muted-foreground hover:text-foreground shrink-0"
            onClick={handleClear}
          />
        ) : (
          <ChevronDown className={cn("h-3 w-3 shrink-0 transition-transform", isOpen && "rotate-180")} />
        )}
      </button>

      {/* Dropdown panel */}
      {isOpen && (
        <div
          className="absolute left-0 top-full mt-1 z-50 w-[280px] bg-popover text-popover-foreground border rounded-md shadow-lg overflow-hidden animate-in fade-in-0 zoom-in-95"
          onKeyDown={handleKeyDown}
        >
          {/* Search input */}
          <div className="p-2 border-b">
            <div className="relative">
              <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <input
                ref={searchInputRef}
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search auras..."
                className="w-full pl-8 pr-2 py-1.5 text-sm bg-transparent border rounded focus:outline-none focus:ring-1 focus:ring-ring"
              />
            </div>
          </div>

          {/* Results */}
          <div className="max-h-[300px] overflow-auto styled-scrollbar">
            <div className="p-1">
              {filteredAuras.length > 0 ? (
                filteredAuras.map((aura) => {
                  const isSelected = selectedSet.has(aura);
                  return (
                    <button
                      key={aura}
                      type="button"
                      onClick={() => handleToggle(aura)}
                      className={cn(
                        "w-full flex items-center gap-2 px-2 py-1.5 text-sm rounded-sm",
                        "hover:bg-accent hover:text-accent-foreground cursor-pointer",
                        isSelected && "bg-accent/50"
                      )}
                    >
                      <div className={cn(
                        "w-4 h-4 rounded border flex items-center justify-center shrink-0",
                        isSelected ? "bg-primary border-primary" : "border-muted-foreground/50"
                      )}>
                        {isSelected && <Check className="h-3 w-3 text-primary-foreground" />}
                      </div>
                      <span className="truncate text-left">{aura}</span>
                    </button>
                  );
                })
              ) : (
                <div className="px-2 py-4 text-sm text-muted-foreground text-center">
                  {auras.length === 0 ? "No auras recorded" : "No matching auras"}
                </div>
              )}
            </div>
          </div>
          
          {/* Count */}
          <div className="px-2 py-1 border-t text-xs text-muted-foreground">
            {filteredAuras.length} of {auras.length} auras
          </div>
        </div>
      )}
    </div>
  );
}
