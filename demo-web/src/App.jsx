import { useEffect, useMemo, useState } from "react";

const API_BASE = import.meta.env.VITE_API_BASE_URL || "http://127.0.0.1:8080";
const PAGE_SIZE = 5;

function paginate(items, page, pageSize) {
  const totalPages = Math.max(1, Math.ceil(items.length / pageSize));
  const safePage = Math.min(Math.max(1, page), totalPages);
  const start = (safePage - 1) * pageSize;
  return {
    pageItems: items.slice(start, start + pageSize),
    totalPages,
    safePage,
  };
}

async function getJSON(path) {
  const res = await fetch(`${API_BASE}${path}`);
  if (!res.ok) {
    throw new Error(`Request failed: ${res.status}`);
  }
  return res.json();
}

async function postJSON(path, body) {
  const res = await fetch(`${API_BASE}${path}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    throw new Error(`Request failed: ${res.status}`);
  }
  return res.json();
}

function formatTimestamp(value) {
  if (!value) {
    return "";
  }

  const match = String(value).match(/^(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2})/);
  if (match) {
    return `${match[1]} ${match[2]}`;
  }

  return String(value).replace("T", " ");
}

function SectionTitle({ children }) {
  return <h2 className="section-title">{children}</h2>;
}

function Pager({ page, totalPages, onPrev, onNext }) {
  return (
    <div className="pager">
      <button onClick={onPrev} disabled={page <= 1}>
        Prev
      </button>
      <span>
        {page} / {totalPages}
      </span>
      <button onClick={onNext} disabled={page >= totalPages}>
        Next
      </button>
    </div>
  );
}

export default function App() {
  const [miners, setMiners] = useState([]);
  const [proofs, setProofs] = useState([]);
  const [files, setFiles] = useState([]);

  const [minersLoading, setMinersLoading] = useState(false);
  const [proofsLoading, setProofsLoading] = useState(false);
  const [filesLoading, setFilesLoading] = useState(false);

  const [minerPage, setMinerPage] = useState(1);
  const [proofPage, setProofPage] = useState(1);

  const [dealFileName, setDealFileName] = useState("");
  const [retrieveFileName, setRetrieveFileName] = useState("");
  const [retrieveOutputName, setRetrieveOutputName] = useState("");

  const [dealLoading, setDealLoading] = useState(false);
  const [retrieveLoading, setRetrieveLoading] = useState(false);

  const [dealMessage, setDealMessage] = useState("");
  const [retrieveMessage, setRetrieveMessage] = useState("");
  const [filesMessage, setFilesMessage] = useState("");

  async function fetchMiners() {
    try {
      setMinersLoading(true);
      const data = await getJSON("/api/miners");
      setMiners(data.miners || []);
    } catch (err) {
      console.error(err);
    } finally {
      setMinersLoading(false);
    }
  }

  async function fetchProofs() {
    try {
      setProofsLoading(true);
      const data = await getJSON("/api/proofs");
      setProofs(data.proofs || []);
    } catch (err) {
      console.error(err);
    } finally {
      setProofsLoading(false);
    }
  }

  async function fetchFiles() {
    try {
      setFilesLoading(true);
      setFilesMessage("");
      const data = await getJSON("/api/files");
      setFiles(data.files || []);
    } catch (err) {
      setFilesMessage(err.message);
    } finally {
      setFilesLoading(false);
    }
  }

  useEffect(() => {
    fetchMiners();
    fetchProofs();

    const timer = setInterval(() => {
      fetchMiners();
      fetchProofs();
    }, 30000);

    return () => clearInterval(timer);
  }, []);

  const minerPagination = useMemo(
    () => paginate(miners, minerPage, PAGE_SIZE),
    [miners, minerPage]
  );

  const proofPagination = useMemo(
    () => paginate(proofs, proofPage, PAGE_SIZE),
    [proofs, proofPage]
  );

  useEffect(() => {
    if (minerPage !== minerPagination.safePage) {
      setMinerPage(minerPagination.safePage);
    }
  }, [minerPage, minerPagination.safePage]);

  useEffect(() => {
    if (proofPage !== proofPagination.safePage) {
      setProofPage(proofPagination.safePage);
    }
  }, [proofPage, proofPagination.safePage]);

  async function handleDeal() {
    if (!dealFileName.trim()) {
      setDealMessage("Please enter a file name.");
      return;
    }

    try {
      setDealLoading(true);
      setDealMessage("Uploading...");
      const data = await postJSON("/api/deal", {
        file_name: dealFileName.trim(),
      });
      setDealMessage(data.message || (data.success ? "Deal finished." : "Deal failed."));
      fetchMiners();
      fetchProofs();
      fetchFiles();
    } catch (err) {
      setDealMessage(`Request failed: ${err.message}`);
    } finally {
      setDealLoading(false);
    }
  }

  async function handleRetrieve() {
    if (!retrieveFileName.trim()) {
      setRetrieveMessage("Please enter a file name.");
      return;
    }

    try {
      setRetrieveLoading(true);
      setRetrieveMessage("Retrieving...");
      const data = await postJSON("/api/retrieve", {
        file_name: retrieveFileName.trim(),
        output_name: retrieveOutputName.trim(),
      });
      if (data.success) {
        setRetrieveMessage(
          data.output_path
            ? `Retrieve finished: ${data.output_path}`
            : "Retrieve finished."
        );
      } else {
        setRetrieveMessage(data.message || "Retrieve failed.");
      }
      fetchProofs();
      fetchFiles();
    } catch (err) {
      setRetrieveMessage(`Request failed: ${err.message}`);
    } finally {
      setRetrieveLoading(false);
    }
  }

  return (
    <div className="app-shell">
      <header className="hero">
        <div className="hero-inner">
          <h1 className="hero-title">TrustDSN Demo System</h1>
          <div className="hero-divider" />
        </div>
      </header>

      <main className="content">
        <section className="info-section">
          <div className="panel miner-panel">
            <SectionTitle>Miner Information</SectionTitle>
            <div className="table-wrap">
              {minersLoading ? (
                <div className="empty-state">Refreshing miner information...</div>
              ) : minerPagination.pageItems.length === 0 ? (
                <div className="empty-state">No miner information available.</div>
              ) : (
                <table className="info-table">
                  <colgroup>
                    <col style={{ width: "24%" }} />
                    <col style={{ width: "13%" }} />
                    <col style={{ width: "19%" }} />
                    <col style={{ width: "21%" }} />
                    <col style={{ width: "23%" }} />
                  </colgroup>
                  <thead>
                    <tr>
                      <th>Node IP</th>
                      <th>Index</th>
                      <th>Storage Power</th>
                      <th>Committed Space</th>
                      <th>User Data Size</th>
                    </tr>
                  </thead>
                  <tbody>
                    {minerPagination.pageItems.map((item, idx) => (
                      <tr key={`${item.node_ip}-${item.index}-${idx}`}>
                        <td>{item.node_ip}</td>
                        <td>{item.index}</td>
                        <td>{item.storage_power}</td>
                        <td>{item.committed_space}</td>
                        <td>{item.user_data_size}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
            <Pager
              page={minerPagination.safePage}
              totalPages={minerPagination.totalPages}
              onPrev={() => setMinerPage((p) => Math.max(1, p - 1))}
              onNext={() =>
                setMinerPage((p) => Math.min(minerPagination.totalPages, p + 1))
              }
            />
          </div>

          <div className="panel proof-panel">
            <SectionTitle>Proof Information</SectionTitle>
            <div className="table-wrap">
              {proofsLoading ? (
                <div className="empty-state">Refreshing proof information...</div>
              ) : proofPagination.pageItems.length === 0 ? (
                <div className="empty-state">No proof information available.</div>
              ) : (
                <table className="info-table proof-table">
                  <colgroup>
                    <col style={{ width: "18%" }} />
                    <col style={{ width: "17%" }} />
                    <col style={{ width: "12%" }} />
                    <col style={{ width: "25%" }} />
                    <col style={{ width: "13%" }} />
                    <col style={{ width: "15%" }} />
                  </colgroup>
                  <thead>
                    <tr>
                      <th>Node IP</th>
                      <th>Proof Type</th>
                      <th>Status</th>
                      <th>Timestamp</th>
                      <th>Generate (s)</th>
                      <th>Verify (ms)</th>
                    </tr>
                  </thead>
                  <tbody>
                    {proofPagination.pageItems.map((item, idx) => (
                      <tr key={`${item.node_ip}-${item.proof_type}-${idx}`}>
                        <td>{item.node_ip}</td>
                        <td>{item.proof_type}</td>
                        <td>{item.status}</td>
                        <td>{formatTimestamp(item.timestamp)}</td>
                        <td>{item.generate_duration_seconds}</td>
                        <td>{item.verify_duration_milliseconds}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
            <Pager
              page={proofPagination.safePage}
              totalPages={proofPagination.totalPages}
              onPrev={() => setProofPage((p) => Math.max(1, p - 1))}
              onNext={() =>
                setProofPage((p) => Math.min(proofPagination.totalPages, p + 1))
              }
            />
          </div>
        </section>

        <section className="action-section">
          <div className="card">
            <SectionTitle>Store File</SectionTitle>
            <input
              type="text"
              placeholder="Enter file name in repo root"
              value={dealFileName}
              onChange={(e) => setDealFileName(e.target.value)}
            />
            <button onClick={handleDeal} disabled={dealLoading}>
              {dealLoading ? "Uploading..." : "Upload File"}
            </button>
            <div className="message-box">{dealMessage || "Ready."}</div>
          </div>

          <div className="card">
            <SectionTitle>Retrieve File</SectionTitle>
            <input
              type="text"
              placeholder="Enter file name"
              value={retrieveFileName}
              onChange={(e) => setRetrieveFileName(e.target.value)}
            />
            <input
              type="text"
              placeholder="Enter output name"
              value={retrieveOutputName}
              onChange={(e) => setRetrieveOutputName(e.target.value)}
            />
            <button onClick={handleRetrieve} disabled={retrieveLoading}>
              {retrieveLoading ? "Retrieving..." : "Retrieve File"}
            </button>
            <div className="message-box">{retrieveMessage || "Ready."}</div>
          </div>

          <div className="card">
            <SectionTitle>Stored Files</SectionTitle>
            <button onClick={fetchFiles} disabled={filesLoading}>
              {filesLoading ? "Loading..." : "Show File Info"}
            </button>
            <div className="file-box">
              {files.length === 0 ? (
                <div className="empty-state">
                  {filesMessage || "No file information loaded."}
                </div>
              ) : (
                files.map((name) => (
                  <div key={name} className="file-item">
                    {name}
                  </div>
                ))
              )}
            </div>
          </div>
        </section>
      </main>
    </div>
  );
}
