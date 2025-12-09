"use client";

import { useState, useEffect } from "react";
import { Plus, Trash2, ExternalLink, TrendingUp, TrendingDown, Minus } from "lucide-react";
import {
  getOrganizationSignals,
  createOrganizationSignal,
  deleteOrganizationSignal,
  getIndividualSignals,
  createIndividualSignal,
  deleteIndividualSignal,
} from "../../api";

const SIGNAL_TYPES = [
  { value: "hiring", label: "Hiring" },
  { value: "expansion", label: "Expansion" },
  { value: "product_launch", label: "Product Launch" },
  { value: "partnership", label: "Partnership" },
  { value: "leadership_change", label: "Leadership Change" },
  { value: "news", label: "News" },
];

const SENTIMENT_OPTIONS = [
  { value: "positive", label: "Positive", icon: TrendingUp, color: "text-green-500" },
  { value: "neutral", label: "Neutral", icon: Minus, color: "text-zinc-500" },
  { value: "negative", label: "Negative", icon: TrendingDown, color: "text-red-500" },
];

const SignalsSection = ({ entityType, entityId }) => {
  const [signals, setSignals] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [showAddForm, setShowAddForm] = useState(false);
  const [error, setError] = useState(null);
  const [newSignal, setNewSignal] = useState({
    signal_type: "news",
    headline: "",
    description: "",
    signal_date: "",
    source: "",
    source_url: "",
    sentiment: "neutral",
  });

  const isOrganization = entityType === "organization";

  useEffect(() => {
    if (entityId) {
      fetchSignals();
    }
  }, [entityId, entityType]);

  const fetchSignals = async () => {
    setIsLoading(true);
    try {
      const data = isOrganization
        ? await getOrganizationSignals(entityId)
        : await getIndividualSignals(entityId);
      setSignals(data || []);
    } catch (err) {
      console.error("Failed to fetch signals:", err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleAddSignal = async () => {
    if (!newSignal.headline.trim()) {
      setError("Headline is required");
      return;
    }

    setError(null);
    try {
      const created = isOrganization
        ? await createOrganizationSignal(entityId, newSignal)
        : await createIndividualSignal(entityId, newSignal);
      setSignals([created, ...signals]);
      setNewSignal({
        signal_type: "news",
        headline: "",
        description: "",
        signal_date: "",
        source: "",
        source_url: "",
        sentiment: "neutral",
      });
      setShowAddForm(false);
    } catch (err) {
      setError(err?.message || "Failed to add signal");
    }
  };

  const handleDeleteSignal = async (signalId) => {
    setError(null);
    try {
      if (isOrganization) {
        await deleteOrganizationSignal(entityId, signalId);
      } else {
        await deleteIndividualSignal(entityId, signalId);
      }
      setSignals(signals.filter((s) => s.id !== signalId));
    } catch (err) {
      setError(err?.message || "Failed to delete signal");
    }
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return null;
    const date = new Date(dateStr);
    return date.toLocaleDateString();
  };

  const getSentimentIcon = (sentiment) => {
    const option = SENTIMENT_OPTIONS.find((o) => o.value === sentiment);
    if (!option) return null;
    const Icon = option.icon;
    return <Icon size={14} className={option.color} />;
  };

  const inputClasses =
    "w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100";

  return (
    <div className="space-y-3">
      {/* Header */}
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium text-zinc-700 dark:text-zinc-300">Signals</span>
        {!showAddForm && (
          <button
            type="button"
            onClick={() => setShowAddForm(true)}
            className="flex items-center gap-1 text-sm text-lime-600 dark:text-lime-400 hover:text-lime-700 dark:hover:text-lime-300">
            <Plus size={16} />
            Add Signal
          </button>
        )}
      </div>

      {/* Error */}
      {error && (
        <div className="p-2 text-sm text-red-600 dark:text-red-400 bg-red-100 dark:bg-red-900/30 rounded">
          {error}
        </div>
      )}

      {/* Add form */}
      {showAddForm && (
        <div className="p-3 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-800/50 space-y-3">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs text-zinc-500 dark:text-zinc-400 mb-1">Type</label>
              <select
                value={newSignal.signal_type}
                onChange={(e) => setNewSignal({ ...newSignal, signal_type: e.target.value })}
                className={inputClasses}>
                {SIGNAL_TYPES.map((type) => (
                  <option key={type.value} value={type.value}>
                    {type.label}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-xs text-zinc-500 dark:text-zinc-400 mb-1">Sentiment</label>
              <select
                value={newSignal.sentiment}
                onChange={(e) => setNewSignal({ ...newSignal, sentiment: e.target.value })}
                className={inputClasses}>
                {SENTIMENT_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div>
            <label className="block text-xs text-zinc-500 dark:text-zinc-400 mb-1">Headline *</label>
            <input
              type="text"
              value={newSignal.headline}
              onChange={(e) => setNewSignal({ ...newSignal, headline: e.target.value })}
              className={inputClasses}
              placeholder="Signal headline"
            />
          </div>

          <div>
            <label className="block text-xs text-zinc-500 dark:text-zinc-400 mb-1">Description</label>
            <textarea
              value={newSignal.description}
              onChange={(e) => setNewSignal({ ...newSignal, description: e.target.value })}
              className={`${inputClasses} resize-none`}
              rows={2}
              placeholder="Additional details..."
            />
          </div>

          <div className="grid grid-cols-3 gap-3">
            <div>
              <label className="block text-xs text-zinc-500 dark:text-zinc-400 mb-1">Date</label>
              <input
                type="date"
                value={newSignal.signal_date}
                onChange={(e) => setNewSignal({ ...newSignal, signal_date: e.target.value })}
                className={inputClasses}
              />
            </div>
            <div>
              <label className="block text-xs text-zinc-500 dark:text-zinc-400 mb-1">Source</label>
              <input
                type="text"
                value={newSignal.source}
                onChange={(e) => setNewSignal({ ...newSignal, source: e.target.value })}
                className={inputClasses}
                placeholder="e.g. LinkedIn"
              />
            </div>
            <div>
              <label className="block text-xs text-zinc-500 dark:text-zinc-400 mb-1">Source URL</label>
              <input
                type="url"
                value={newSignal.source_url}
                onChange={(e) => setNewSignal({ ...newSignal, source_url: e.target.value })}
                className={inputClasses}
                placeholder="https://..."
              />
            </div>
          </div>

          <div className="flex gap-2">
            <button
              type="button"
              onClick={handleAddSignal}
              className="px-3 py-1 text-sm bg-lime-600 text-white rounded hover:bg-lime-700">
              Add
            </button>
            <button
              type="button"
              onClick={() => setShowAddForm(false)}
              className="px-3 py-1 text-sm text-zinc-600 dark:text-zinc-400 hover:text-zinc-800 dark:hover:text-zinc-200">
              Cancel
            </button>
          </div>
        </div>
      )}

      {/* Loading */}
      {isLoading && <p className="text-sm text-zinc-500">Loading signals...</p>}

      {/* Signals list */}
      {!isLoading && signals.length > 0 && (
        <div className="space-y-2">
          {signals.map((signal) => (
            <div
              key={signal.id}
              className="p-3 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-white dark:bg-zinc-800">
              <div className="flex items-start justify-between gap-2">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="inline-flex items-center px-2 py-0.5 text-xs font-medium rounded bg-zinc-100 dark:bg-zinc-700 text-zinc-600 dark:text-zinc-300">
                      {SIGNAL_TYPES.find((t) => t.value === signal.signal_type)?.label || signal.signal_type}
                    </span>
                    {signal.sentiment && getSentimentIcon(signal.sentiment)}
                    {signal.signal_date && (
                      <span className="text-xs text-zinc-500 dark:text-zinc-400">
                        {formatDate(signal.signal_date)}
                      </span>
                    )}
                  </div>
                  <h4 className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
                    {signal.headline}
                  </h4>
                  {signal.description && (
                    <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-400 line-clamp-2">
                      {signal.description}
                    </p>
                  )}
                  {(signal.source || signal.source_url) && (
                    <div className="mt-1 flex items-center gap-2 text-xs text-zinc-500 dark:text-zinc-400">
                      {signal.source && <span>{signal.source}</span>}
                      {signal.source_url && (
                        <a
                          href={signal.source_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="inline-flex items-center gap-1 text-lime-600 dark:text-lime-400 hover:underline">
                          <ExternalLink size={12} />
                          Link
                        </a>
                      )}
                    </div>
                  )}
                </div>
                <button
                  type="button"
                  onClick={() => handleDeleteSignal(signal.id)}
                  className="p-1 text-zinc-400 hover:text-red-500"
                  title="Delete signal">
                  <Trash2 size={16} />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Empty state */}
      {!isLoading && !showAddForm && signals.length === 0 && (
        <p className="text-sm text-zinc-500 dark:text-zinc-400">No signals recorded yet.</p>
      )}
    </div>
  );
};

export default SignalsSection;
